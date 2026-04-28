// Программа seed заполняет оркестратор тестовыми данными.
//
// Назначение:
//   - Подготовка окружения для ручного тестирования
//   - Подготовка данных для нагрузочного тестирования
//   - Быстрое развёртывание демо-стенда
//
// Использование:
//
//	go run ./cmd/seed
//	go run ./cmd/seed --config config/local.yaml
//	go run ./cmd/seed --nodes 10 --instances 50
//	go run ./cmd/seed --clean  # очистить перед заполнением
//	go run ./cmd/seed --db-host localhost --db-port 5432 --kv-addr localhost:6379  # локальный запуск
package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Ключи KV-хранилища (должны совпадать с valkey/instance_state_store.go)
const (
	keyInstanceStatus = "inst:st:"
	keyInstanceCount  = "inst:pc:"
	keyInstanceUsage  = "inst:us:"
)

// SeedConfig — параметры seed-скрипта.
type SeedConfig struct {
	Nodes     int
	Builds    int
	Instances int
	Clean     bool
	Quiet     bool
	// Переопределение подключения (для запуска вне Docker)
	DBHost string
	DBPort int
	KVAddr string
}

func main() {
	cfg := parseFlags()

	seed := NewSeeder(cfg)
	if err := seed.Run(context.Background()); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
}

func parseFlags() SeedConfig {
	var cfg SeedConfig

	// Используем отдельный FlagSet чтобы не конфликтовать с config.MustLoad()
	fs := flag.NewFlagSet("seed", flag.ContinueOnError)
	fs.IntVar(&cfg.Nodes, "nodes", 0, "количество тестовых нод")
	fs.IntVar(&cfg.Builds, "builds", 0, "количество тестовых билдов (на игру)")
	fs.IntVar(&cfg.Instances, "instances", 0, "количество тестовых инстансов")
	fs.BoolVar(&cfg.Clean, "clean", false, "очистить таблицы перед заполнением")
	fs.BoolVar(&cfg.Quiet, "quiet", false, "минимизировать вывод")
	fs.StringVar(&cfg.DBHost, "db-host", "", "переопределить host PostgreSQL (для локального запуска)")
	fs.IntVar(&cfg.DBPort, "db-port", 0, "переопределить port PostgreSQL (для локального запуска)")
	fs.StringVar(&cfg.KVAddr, "kv-addr", "", "переопределить адрес Valkey (для локального запуска)")

	// Игнорируем ошибки парсинга - флаги могут быть не указаны
	_ = fs.Parse(os.Args[1:])

	// Читаем из переменных окружения если флаги не указаны
	if cfg.Nodes == 0 {
		if v := os.Getenv("SEED_NODES"); v != "" {
			_, _ = fmt.Sscanf(v, "%d", &cfg.Nodes)
		}
	}
	if cfg.Nodes == 0 {
		cfg.Nodes = 3 // default
	}

	if cfg.Builds == 0 {
		if v := os.Getenv("SEED_BUILDS"); v != "" {
			_, _ = fmt.Sscanf(v, "%d", &cfg.Builds)
		}
	}
	if cfg.Builds == 0 {
		cfg.Builds = 2 // default
	}

	if cfg.Instances == 0 {
		if v := os.Getenv("SEED_INSTANCES"); v != "" {
			_, _ = fmt.Sscanf(v, "%d", &cfg.Instances)
		}
	}
	if cfg.Instances == 0 {
		cfg.Instances = 8 // default
	}

	if os.Getenv("SEED_CLEAN") == "1" {
		cfg.Clean = true
	}

	if cfg.DBHost == "" {
		cfg.DBHost = os.Getenv("SEED_DB_HOST")
	}
	if cfg.DBPort == 0 {
		if v := os.Getenv("SEED_DB_PORT"); v != "" {
			_, _ = fmt.Sscanf(v, "%d", &cfg.DBPort)
		}
	}
	if cfg.KVAddr == "" {
		cfg.KVAddr = os.Getenv("SEED_KV_ADDR")
	}

	return cfg
}

// Seeder управляет процессом заполнения данными.
type Seeder struct {
	cfg     SeedConfig
	appCfg  *config.Config
	pgPool  *pgxpool.Pool
	kv      *redis.Client
	verbose bool
}

// NewSeeder создаёт seeder.
func NewSeeder(cfg SeedConfig) *Seeder {
	return &Seeder{cfg: cfg, verbose: !cfg.Quiet}
}

// Run выполняет заполнение данными.
func (s *Seeder) Run(ctx context.Context) error {
	s.log("[INFO] Заполнение оркестратора тестовыми данными")

	// 1. Загружаем конфигурацию
	if err := s.loadConfig(); err != nil {
		return fmt.Errorf("загрузка конфигурации: %w", err)
	}

	// 2. Подключаемся к БД
	if err := s.connectDB(ctx); err != nil {
		return fmt.Errorf("подключение к БД: %w", err)
	}
	defer s.pgPool.Close()

	// 3. Подключаемся к KV
	if err := s.connectKV(ctx); err != nil {
		return fmt.Errorf("подключение к KV: %w", err)
	}
	defer func() { _ = s.kv.Close() }()

	// 4. Очистка (опционально)
	if s.cfg.Clean {
		if err := s.cleanData(ctx); err != nil {
			return fmt.Errorf("очистка данных: %w", err)
		}
	}

	// 5. Заполняем данные
	if err := s.seedNodes(ctx); err != nil {
		return fmt.Errorf("заполнение нод: %w", err)
	}
	if err := s.seedBuilds(ctx); err != nil {
		return fmt.Errorf("заполнение билдов: %w", err)
	}
	if err := s.seedInstances(ctx); err != nil {
		return fmt.Errorf("заполнение инстансов: %w", err)
	}
	if err := s.seedPlayerCounts(ctx); err != nil {
		return fmt.Errorf("заполнение player counts: %w", err)
	}

	s.log("[OK] Заполнение завершено успешно")
	s.printSummary()
	return nil
}

func (s *Seeder) loadConfig() error {
	// Конфигурация загружается через CONFIG_PATH env
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/local.yaml"
	}

	// Устанавливаем env для cleanenv
	if err := os.Setenv("CONFIG_PATH", configPath); err != nil {
		return err
	}

	s.appCfg = config.MustLoad()
	s.log("  [CONFIG] Конфигурация загружена: %s", configPath)
	return nil
}

func (s *Seeder) connectDB(ctx context.Context) error {
	// Переопределяем host/port если указаны (для локального запуска)
	host := s.appCfg.DB.Host
	port := s.appCfg.DB.Port

	if s.cfg.DBHost != "" {
		host = s.cfg.DBHost
	}
	if s.cfg.DBPort != 0 {
		port = s.cfg.DBPort
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		s.appCfg.DB.User,
		s.appCfg.DB.Password,
		host,
		port,
		s.appCfg.DB.Name,
		s.appCfg.DB.SSLMode,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	if err := pool.Ping(ctx); err != nil {
		return err
	}
	s.pgPool = pool
	s.log("  [DB] PostgreSQL: %s:%d/%s", host, port, s.appCfg.DB.Name)
	return nil
}

func (s *Seeder) connectKV(ctx context.Context) error {
	// Переопределяем адрес если указан (для локального запуска)
	addr := s.appCfg.KV.Addr
	if s.cfg.KVAddr != "" {
		addr = s.cfg.KVAddr
	}

	s.kv = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: s.appCfg.KV.Password,
		DB:       s.appCfg.KV.DB,
	})
	if err := s.kv.Ping(ctx).Err(); err != nil {
		return err
	}
	s.log("  [KV] Valkey: %s", addr)
	return nil
}

func (s *Seeder) cleanData(ctx context.Context) error {
	s.log("  [CLEAN] Очистка таблиц...")
	const q = `TRUNCATE instances, server_builds, nodes RESTART IDENTITY CASCADE`
	_, err := s.pgPool.Exec(ctx, q)
	if err != nil {
		return err
	}

	// Очищаем KV
	if err := s.kv.FlushDB(ctx).Err(); err != nil {
		return err
	}

	s.log("    [OK] Таблицы очищены")
	return nil
}

func (s *Seeder) seedNodes(ctx context.Context) error {
	s.log("  [NODES] Создание %d нод...", s.cfg.Nodes)

	repo := postgres.NewNodeRepo(s.pgPool)
	now := time.Now()

	for i := 1; i <= s.cfg.Nodes; i++ {
		// Test credentials for seed data (not real secrets)
		const seedToken = "seed-api-token" //nolint:gosec // test data only

		node := &domain.Node{
			ID:           int64(i),
			Address:      fmt.Sprintf("seed-node-%d:44044", i),
			TokenHash:    hashToken("seed-token"),
			APIToken:     seedToken,
			Region:       "seed-region",
			Status:       domain.NodeStatusOnline,
			CPUCores:     uint32(4 + (i%4)*2),        // 4-10 cores
			TotalMemory:  uint64(8+(i%3)*8) << 30,    // 8-24 GB
			TotalDisk:    uint64(100+(i%5)*50) << 30, // 100-300 GB
			AgentVersion: "0.0.1-seed",
			LastPingAt:   now,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := repo.Create(ctx, node); err != nil {
			return err
		}
	}

	s.log("    [OK] Создано нод: %d", s.cfg.Nodes)
	return nil
}

func (s *Seeder) seedBuilds(ctx context.Context) error {
	s.log("  [BUILDS] Создание %d билдов...", s.cfg.Builds)

	store := postgres.NewBuildStorage(s.pgPool)
	now := time.Now()

	// Билды для игр 1 и 2
	for gameID := int64(1); gameID <= 2; gameID++ {
		for i := 1; i <= s.cfg.Builds; i++ {
			build := &domain.ServerBuild{
				ID:           int64((gameID-1)*int64(s.cfg.Builds) + int64(i)),
				GameID:       gameID,
				UploadedBy:   0, // system
				Version:      fmt.Sprintf("1.%d.0", i-1),
				ImageTag:     fmt.Sprintf("welwise/game-%d:1.%d.0", gameID, i-1),
				Protocol:     domain.ProtocolTCP,
				InternalPort: 8080,
				MaxPlayers:   uint32(50 + gameID*50),
				FileURL:      fmt.Sprintf("/builds/game-%d-1.%d.0.tar", gameID, i-1),
				FileSize:     1024 * 1024, // 1 MB dummy
				CreatedAt:    now,
			}

			if err := store.Create(ctx, build); err != nil {
				return err
			}
		}
	}

	s.log("    [OK] Создано билдов: %d", s.cfg.Builds*2)
	return nil
}

func (s *Seeder) seedInstances(ctx context.Context) error {
	s.log("  [INSTANCES] Создание %d инстансов...", s.cfg.Instances)

	repo := postgres.NewInstanceRepo(s.pgPool)
	now := time.Now()

	// Равномерно распределяем инстансы по играм и нодам
	for i := 1; i <= s.cfg.Instances; i++ {
		gameID := int64((i-1)%2 + 1) // игра 1 или 2
		nodeID := int64((i-1)%s.cfg.Nodes + 1)
		buildID := int64((gameID-1)*int64(s.cfg.Builds) + 1) // первый билд игры

		instance := &domain.Instance{
			ID:            int64(i),
			NodeID:        nodeID,
			ServerBuildID: buildID,
			GameID:        gameID,
			Name:          fmt.Sprintf("seed-inst-%d", i),
			BuildVersion:  "1.0.0",
			Protocol:      domain.ProtocolTCP,
			HostPort:      uint32(7000 + i),
			InternalPort:  8080,
			Status:        domain.InstanceStatusRunning,
			MaxPlayers:    uint32(50 + gameID*50),
			ServerAddress: fmt.Sprintf("seed-node-%d.example.com", nodeID),
			StartedAt:     now,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := repo.Create(ctx, instance); err != nil {
			return err
		}
	}

	s.log("    [OK] Создано инстансов: %d", s.cfg.Instances)
	return nil
}

func (s *Seeder) seedPlayerCounts(ctx context.Context) error {
	s.log("  [KV] Заполнение player counts...")

	// Ключи в Valkey: inst:pc:<instanceID>
	// Значения: количество игроков (случайное, но < max_players)
	kvPairs := make(map[string]interface{})

	for i := 1; i <= s.cfg.Instances; i++ {
		gameID := (i-1)%2 + 1
		maxPlayers := 50 + gameID*50
		playerCount := (i * 7) % maxPlayers // детерминированное "случайное" число

		key := fmt.Sprintf("%s%d", keyInstanceCount, i)
		kvPairs[key] = playerCount
	}

	if err := s.kv.MSet(ctx, kvPairs).Err(); err != nil {
		return err
	}

	s.log("    [OK] Записей player counts: %d", len(kvPairs))
	return nil
}

func (s *Seeder) printSummary() {
	fmt.Println()
	fmt.Println("===================================================")
	fmt.Println("  ИТОГОВАЯ СВОДКА")
	fmt.Println("===================================================")
	fmt.Printf("  Нод:       %d\n", s.cfg.Nodes)
	fmt.Printf("  Билдов:    %d (на игру)\n", s.cfg.Builds)
	fmt.Printf("  Инстансов: %d\n", s.cfg.Instances)
	fmt.Println()
	fmt.Println("  Данные готовы для:")
	fmt.Println("    - Ручного тестирования UI")
	fmt.Println("    - Нагрузочного тестирования")
	fmt.Println("    - Демонстрации функционала")
	fmt.Println("===================================================")
}

func (s *Seeder) log(format string, args ...any) {
	if s.verbose {
		fmt.Printf(format+"\n", args...)
	}
}

func hashToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}
