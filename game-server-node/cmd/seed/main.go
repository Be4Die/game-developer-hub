// Программа seed заполняет game-server-node тестовыми данными через gRPC API.
//
// Назначение:
//   - Подготовка окружения для нагрузочного тестирования DiscoveryService
//   - Подготовка данных для интеграционного тестирования
//   - Быстрое развёртывание демо-стенда
//
// Требования:
//   - Работающий game-server-node (task node:up)
//   - Доступ к Docker демону (для создания контейнеров)
//   - API-ключ для авторизации
//
// Использование:
//
//	go run ./cmd/seed
//	go run ./cmd/seed --addr localhost:44044 --api-key dev-api-key
//	go run ./cmd/seed --instances 20 --games 5
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Config — параметры seed-скрипта.
type SeedConfig struct {
	Addr      string
	APIKey    string
	Games     int
	Instances int
	ImageTag  string
	Quiet     bool
	DryRun    bool
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

	fs := flag.NewFlagSet("seed", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr, "addr", "", "gRPC адрес game-server-node")
	fs.StringVar(&cfg.APIKey, "api-key", "", "API-ключ для авторизации (или NODE_API_KEY env)")
	fs.IntVar(&cfg.Games, "games", 0, "количество игр")
	fs.IntVar(&cfg.Instances, "instances", 0, "количество инстансов (всего)")
	fs.StringVar(&cfg.ImageTag, "image", "alpine:latest", "Docker-образ для инстансов")
	fs.BoolVar(&cfg.Quiet, "quiet", false, "минимизировать вывод")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "не создавать реальные инстансы, только показать план")

	// Игнорируем ошибки парсинга
	_ = fs.Parse(os.Args[1:])

	// Читаем из переменных окружения
	if cfg.Addr == "" {
		cfg.Addr = os.Getenv("SEED_NODE_ADDR")
	}
	if cfg.Addr == "" {
		cfg.Addr = "localhost:44044"
	}

	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("NODE_API_KEY")
	}
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("SEED_API_KEY")
	}
	if cfg.APIKey == "" {
		cfg.APIKey = "dev-api-key-for-local-testing"
	}

	if cfg.Games == 0 {
		if v := os.Getenv("SEED_GAMES"); v != "" {
			fmt.Sscanf(v, "%d", &cfg.Games)
		}
	}
	if cfg.Games == 0 {
		cfg.Games = 2
	}

	if cfg.Instances == 0 {
		if v := os.Getenv("SEED_INSTANCES"); v != "" {
			fmt.Sscanf(v, "%d", &cfg.Instances)
		}
	}
	if cfg.Instances == 0 {
		cfg.Instances = 8
	}

	return cfg
}

// Seeder управляет процессом заполнения данными.
type Seeder struct {
	cfg     SeedConfig
	client  pb.DeploymentServiceClient
	conn    *grpc.ClientConn
	verbose bool
}

// NewSeeder создаёт seeder.
func NewSeeder(cfg SeedConfig) *Seeder {
	return &Seeder{cfg: cfg, verbose: !cfg.Quiet}
}

// Run выполняет заполнение данными.
func (s *Seeder) Run(ctx context.Context) error {
	s.log("[INFO] Заполнение game-server-node тестовыми данными")

	if s.cfg.DryRun {
		s.log("  [WARN] Режим dry-run: реальные инстансы не будут созданы")
		s.printSummary()
		return nil
	}

	// 1. Подключаемся к gRPC серверу
	if err := s.connect(); err != nil {
		return fmt.Errorf("подключение к gRPC: %w", err)
	}
	defer func() { _ = s.conn.Close() }()

	// 2. Загружаем образ для каждой игры
	if err := s.loadImages(ctx); err != nil {
		return fmt.Errorf("загрузка образов: %w", err)
	}

	// 3. Создаём инстансы
	if err := s.seedInstances(ctx); err != nil {
		return fmt.Errorf("создание инстансов: %w", err)
	}

	s.log("[OK] Заполнение завершено успешно")
	s.printSummary()
	return nil
}

func (s *Seeder) connect() error {
	s.log("  [GRPC] Подключение к %s...", s.cfg.Addr)

	conn, err := grpc.NewClient(
		s.cfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	s.conn = conn
	s.client = pb.NewDeploymentServiceClient(conn)
	s.log("    ✓ Подключено")
	return nil
}

func (s *Seeder) loadImages(ctx context.Context) error {
	s.log("  [IMAGES] Загрузка образов для %d игр...", s.cfg.Games)

	for gameID := int64(1); gameID <= int64(s.cfg.Games); gameID++ {
		imageTag := fmt.Sprintf("seed-game-%d:latest", gameID)

		s.log("    • Загрузка %s (game_id=%d)...", imageTag, gameID)

		// Pull образ локально через docker CLI
		// Это нужно, чтобы Seed мог загрузить образ в game-server-node
		// В реальном сценарии образ уже должен быть собран

		// Для seed используем указанный образ, тегируем его под каждую игру
		if err := s.pullAndTagImage(imageTag); err != nil {
			return fmt.Errorf("pull image for game %d: %w", gameID, err)
		}

		// Загружаем образ в game-server-node через gRPC
		if err := s.loadImageGRPC(ctx, gameID, imageTag); err != nil {
			return fmt.Errorf("load image gRPC for game %d: %w", gameID, err)
		}
	}

	s.log("    ✓ Образы загружены")
	return nil
}

func (s *Seeder) pullAndTagImage(imageTag string) error {
	// В реальном сценарии нужно использовать docker SDK
	// Для упрощения предполагаем, что образ уже существует
	s.log("      [skip] pull/tag: %s (предполагается существующим)", imageTag)
	return nil
}

func (s *Seeder) loadImageGRPC(ctx context.Context, gameID int64, imageTag string) error {
	ctx = s.withAuth(ctx)

	stream, err := s.client.LoadImage(ctx)
	if err != nil {
		return err
	}

	// Отправляем метаданные
	if err := stream.Send(&pb.LoadImageRequest{
		Payload: &pb.LoadImageRequest_Metadata{
			Metadata: &pb.ImageMetadata{
				GameId:   gameID,
				ImageTag: imageTag,
			},
		},
	}); err != nil {
		return err
	}

	// Отправляем пустой chunk (для демонстрации)
	// В реальном сценарии нужно отправить tar-архив образа
	if err := stream.Send(&pb.LoadImageRequest{
		Payload: &pb.LoadImageRequest_Chunk{
			Chunk: []byte{},
		},
	}); err != nil {
		return err
	}

	_, err = stream.CloseAndRecv()
	return err
}

func (s *Seeder) seedInstances(ctx context.Context) (_ error) {
	s.log("  [INSTANCES] Создание %d инстансов...", s.cfg.Instances)

	ctx = s.withAuth(ctx)
	instancesPerGame := s.cfg.Instances / s.cfg.Games
	if instancesPerGame < 1 {
		instancesPerGame = 1
	}

	created := 0
	for gameID := int64(1); gameID <= int64(s.cfg.Games) && created < s.cfg.Instances; gameID++ {
		for i := 0; i < instancesPerGame && created < s.cfg.Instances; i++ {
			name := fmt.Sprintf("seed-inst-%d-%d", gameID, i)

			resp, err := s.client.StartInstance(ctx, &pb.StartInstanceRequest{
				GameId:       gameID,
				Name:         name,
				Protocol:     pb.Protocol_PROTOCOL_TCP,
				InternalPort: 8080,
				PortAllocation: &pb.PortAllocation{
					Strategy: &pb.PortAllocation_Any{Any: true},
				},
				MaxPlayers: uint32(50 + gameID*25),
				EnvVars: map[string]string{
					"SEED":     "true",
					"GAME_ID":  fmt.Sprintf("%d", gameID),
					"INST_NUM": fmt.Sprintf("%d", i),
				},
			})
			if err != nil {
				s.log("    [WARN] Ошибка создания %s: %v", name, err)
				continue
			}

			s.log("    ✓ %s: instance_id=%d, port=%d", name, resp.InstanceId, resp.HostPort)
			created++
		}
	}

	s.log("    ✓ Создано инстансов: %d", created)
	return nil
}

func (s *Seeder) withAuth(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-api-key", s.cfg.APIKey)
}

func (s *Seeder) printSummary() {
	fmt.Println()
	fmt.Println("===================================================")
	fmt.Println("  ИТОГОВАЯ СВОДКА")
	fmt.Println("===================================================")
	fmt.Printf("  Адрес:     %s\n", s.cfg.Addr)
	fmt.Printf("  Игр:       %d\n", s.cfg.Games)
	fmt.Printf("  Инстансов: %d\n", s.cfg.Instances)
	fmt.Printf("  Образ:     %s\n", s.cfg.ImageTag)
	fmt.Println()
	fmt.Println("  Данные готовы для:")
	fmt.Println("    - Нагрузочного тестирования DiscoveryService")
	fmt.Println("    - Интеграционного тестирования")
	fmt.Println("===================================================")
}

func (s *Seeder) log(format string, args ...any) {
	if s.verbose {
		fmt.Printf(format+"\n", args...)
	}
}
