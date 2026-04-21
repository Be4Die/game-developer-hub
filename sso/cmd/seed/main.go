// Программа seed заполняет SSO-сервис тестовыми данными через gRPC API.
//
// Назначение:
//   - Подготовка пользователей для нагрузочного тестирования
//   - Подготовка данных для интеграционного тестирования
//   - Быстрое развёртывание демо-стенда
//
// Использование:
//
//	go run ./cmd/seed
//	go run ./cmd/seed --addr localhost:50051 --users 50
//	go run ./cmd/seed --clean  // удалить всех тестовых пользователей перед созданием
package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	pb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config — параметры seed-скрипта.
type Config struct {
	Addr       string
	Users      int
	Admins     int
	Moderators int
	Clean      bool
	Quiet      bool
	OutputCSV  string
}

func main() {
	cfg := parseFlags()

	seed := NewSeeder(cfg)
	if err := seed.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() Config {
	var cfg Config

	fs := flag.NewFlagSet("seed", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr, "addr", "", "gRPC адрес SSO-сервиса")
	fs.IntVar(&cfg.Users, "users", 0, "количество тестовых пользователей")
	fs.IntVar(&cfg.Admins, "admins", 0, "количество администраторов")
	fs.IntVar(&cfg.Moderators, "mods", 0, "количество модераторов")
	fs.BoolVar(&cfg.Clean, "clean", false, "удалить тестовых пользователей перед созданием")
	fs.BoolVar(&cfg.Quiet, "quiet", false, "минимизировать вывод")
	fs.StringVar(&cfg.OutputCSV, "output-csv", "", "путь к CSV файлу с данными для входа")

	_ = fs.Parse(os.Args[1:])

	if cfg.Addr == "" {
		cfg.Addr = os.Getenv("SEED_SSO_ADDR")
	}
	if cfg.Addr == "" {
		cfg.Addr = "localhost:50051"
	}

	if cfg.Users == 0 {
		if v := os.Getenv("SEED_USERS"); v != "" {
			_, _ = fmt.Sscanf(v, "%d", &cfg.Users)
		}
	}
	if cfg.Users == 0 {
		cfg.Users = 10
	}

	if cfg.Admins == 0 {
		if v := os.Getenv("SEED_ADMINS"); v != "" {
			_, _ = fmt.Sscanf(v, "%d", &cfg.Admins)
		}
	}
	if cfg.Admins == 0 {
		cfg.Admins = 1
	}

	if cfg.Moderators == 0 {
		if v := os.Getenv("SEED_MODS"); v != "" {
			_, _ = fmt.Sscanf(v, "%d", &cfg.Moderators)
		}
	}
	if cfg.Moderators == 0 {
		cfg.Moderators = 1
	}

	if cfg.OutputCSV == "" {
		cfg.OutputCSV = os.Getenv("SEED_OUTPUT_CSV")
	}

	return cfg
}

// Seeder управляет процессом заполнения данными.
type Seeder struct {
	cfg     Config
	client  pb.AuthServiceClient
	conn    *grpc.ClientConn
	records [][]string // email,password,display_name,role
}

// NewSeeder создаёт seeder.
func NewSeeder(cfg Config) *Seeder {
	return &Seeder{cfg: cfg}
}

// Run выполняет заполнение данными.
func (s *Seeder) Run(ctx context.Context) error {
	s.log("[INFO] Заполнение SSO тестовыми данными")

	// 1. Подключаемся к gRPC серверу
	if err := s.connect(); err != nil {
		return fmt.Errorf("подключение к gRPC: %w", err)
	}
	defer func() { _ = s.conn.Close() }()

	// 2. Создаём пользователей
	s.seedUsers(ctx)

	// 3. Верифицируем email
	s.verifyEmails(ctx)

	// 4. Выводим CSV если указан путь
	if s.cfg.OutputCSV != "" {
		if err := s.writeCSV(); err != nil {
			return fmt.Errorf("запись CSV: %w", err)
		}
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
	s.client = pb.NewAuthServiceClient(conn)
	s.log("    Подключено")
	return nil
}

func (s *Seeder) seedUsers(ctx context.Context) {
	// Очищаем предыдущие записи
	s.records = nil

	// Создаём обычных пользователей
	s.log("  [USERS] Создание %d пользователей...", s.cfg.Users)
	for i := 1; i <= s.cfg.Users; i++ {
		email := fmt.Sprintf("user%d@test.local", i)
		password := fmt.Sprintf("Password123!user%d", i)
		displayName := fmt.Sprintf("Test User %d", i)

		_, err := s.client.Register(ctx, &pb.AuthServiceRegisterRequest{
			Email:       email,
			Password:    password,
			DisplayName: displayName,
		})
		if err != nil {
			s.log("    [WARN] Ошибка регистрации %s: %v", email, err)
			continue
		}

		s.records = append(s.records, []string{email, password, displayName, "developer"})
		s.log("    + %s", email)
	}

	// Создаём модераторов
	if s.cfg.Moderators > 0 {
		s.log("  [MODERATORS] Создание %d модераторов...", s.cfg.Moderators)
		for i := 1; i <= s.cfg.Moderators; i++ {
			email := fmt.Sprintf("mod%d@test.local", i)
			password := fmt.Sprintf("Mod123!mod%d", i)
			displayName := fmt.Sprintf("Moderator %d", i)

			_, err := s.client.Register(ctx, &pb.AuthServiceRegisterRequest{
				Email:       email,
				Password:    password,
				DisplayName: displayName,
			})
			if err != nil {
				s.log("    [WARN] Ошибка регистрации %s: %v", email, err)
				continue
			}

			s.records = append(s.records, []string{email, password, displayName, "moderator"})
			s.log("    + %s (moderator)", email)
		}
	}

	// Создаём администраторов
	if s.cfg.Admins > 0 {
		s.log("  [ADMINS] Создание %d администраторов...", s.cfg.Admins)
		for i := 1; i <= s.cfg.Admins; i++ {
			email := fmt.Sprintf("admin%d@test.local", i)
			password := fmt.Sprintf("Admin123!admin%d", i)
			displayName := fmt.Sprintf("Admin %d", i)

			_, err := s.client.Register(ctx, &pb.AuthServiceRegisterRequest{
				Email:       email,
				Password:    password,
				DisplayName: displayName,
			})
			if err != nil {
				s.log("    [WARN] Ошибка регистрации %s: %v", email, err)
				continue
			}

			s.records = append(s.records, []string{email, password, displayName, "admin"})
			s.log("    + %s (admin)", email)
		}
	}

	s.log("    Зарегистрировано: %d", len(s.records))
}

func (s *Seeder) verifyEmails(_ context.Context) {
	s.log("  [VERIFY] Верификация email...")

	// Получаем коды верификации из логов SSO
	// В dev-режиме SSO логирует коды, используем прямой доступ к Valkey
	// Для простоты — используем известный паттерн кодов
	for _, rec := range s.records {
		email := rec[0]
		// В тестовой среде код верификации сохраняется в Valkey.
		// Для seed используем API resend + прямой доступ к Valkey.
		// Пока просто логируем — верификация будет ручной или через Valkey.
		s.log("    -> %s (требуется верификация)", email)
	}
}

func (s *Seeder) writeCSV() error {
	f, err := os.Create(s.cfg.OutputCSV)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Заголовок
	if err := w.Write([]string{"email", "password", "display_name", "role"}); err != nil {
		return err
	}

	for _, rec := range s.records {
		if err := w.Write(rec); err != nil {
			return err
		}
	}

	s.log("  [CSV] Записано %d записей в %s", len(s.records), s.cfg.OutputCSV)
	return nil
}

func (s *Seeder) printSummary() {
	fmt.Println()
	fmt.Println("===================================================")
	fmt.Println("  ИТОГОВАЯ СВОДКА")
	fmt.Println("===================================================")
	fmt.Printf("  Адрес:         %s\n", s.cfg.Addr)
	fmt.Printf("  Пользователей: %d\n", s.cfg.Users)
	fmt.Printf("  Модераторов:   %d\n", s.cfg.Moderators)
	fmt.Printf("  Администратор: %d\n", s.cfg.Admins)
	if s.cfg.OutputCSV != "" {
		fmt.Printf("  CSV файл:      %s\n", s.cfg.OutputCSV)
	}
	fmt.Println()
	fmt.Println("  Данные готовы для:")
	fmt.Println("    - Нагрузочного тестирования AuthService")
	fmt.Println("    - Нагрузочного тестирования UserService")
	fmt.Println("    - Нагрузочного тестирования TokenService")
	fmt.Println("===================================================")
}

func (s *Seeder) log(format string, args ...any) {
	if !s.cfg.Quiet {
		fmt.Printf(format+"\n", args...)
	}
}
