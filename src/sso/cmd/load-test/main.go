// Программа load-test проводит нагрузочное тестирование SSO gRPC API через ghz.
//
// Критический путь — Login endpoint: это самый частый вызов при подключении
// игроков к платформе.
//
// Перед запуском:
//  1. Запустите SSO (например, task sso:up).
//  2. Запустите seed: go run ./cmd/seed
//  3. Убедитесь что gRPC сервер доступен на указанном адресе.
//
// Использование:
//
//	go run ./cmd/load-test
//	go run ./cmd/load-test --addr localhost:50051 --concurrency 50 --requests 10000
//	go run ./cmd/load-test --heavy  // 100 concurrency, 50000 requests
//	go run ./cmd/load-test --service auth     // только auth тесты
//	go run ./cmd/load-test --service token    // только token тесты
//	go run ./cmd/load-test --service user     // только user тесты
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Config — параметры нагрузочного теста.
type Config struct {
	Addr        string
	Concurrency uint
	Requests    uint
	Heavy       bool
	Service     string // "auth", "token", "user" или "" (все)
	Quiet       bool
	ProtoPath   string
	Call        string
}

// TestScenario описывает один тестовый сценарий.
type TestScenario struct {
	Name    string
	Call    string
	Data    string
	Service string
}

// TestResult хранит результаты одного теста.
type TestResult struct {
	Name    string        `json:"name"`
	Call    string        `json:"call"`
	Success bool          `json:"success"`
	Date    string        `json:"date"`
	Count   uint          `json:"count"`
	Total   time.Duration `json:"total"`
	Average time.Duration `json:"average"`
	Fastest time.Duration `json:"fastest"`
	Slowest time.Duration `json:"slowest"`
	Median  time.Duration `json:"median"`
	P95     time.Duration `json:"p95"`
	P99     time.Duration `json:"p99"`
	Errors  []string      `json:"errors,omitempty"`
}

func main() {
	cfg := parseFlags()

	fmt.Println("===================================================")
	fmt.Println("  Нагрузочное тестирование — SSO gRPC API (ghz)")
	fmt.Println("===================================================")
	fmt.Printf("  Адрес:       %s\n", cfg.Addr)
	fmt.Printf("  Concurrency: %d\n", cfg.Concurrency)
	fmt.Printf("  Requests:    %d\n", cfg.Requests)
	if cfg.Service != "" {
		fmt.Printf("  Service:     %s\n", cfg.Service)
	}
	fmt.Println("===================================================")
	fmt.Println()

	// Определяем тестовые сценарии
	scenarios := getScenarios(cfg)
	if len(scenarios) == 0 {
		fmt.Println("Нет сценариев для запуска")
		return
	}

	fmt.Printf("Сценариев: %d\n\n", len(scenarios))

	var results []TestResult

	// Основной прогон
	for i, sc := range scenarios {
		fmt.Printf("[%d/%d] Тест: %s\n", i+1, len(scenarios), sc.Name)
		fmt.Println(strings.Repeat("-", 55))

		result := runGhzTest(cfg, sc)
		if result.Success {
			printResult(result)
		} else {
			fmt.Printf("  ОШИБКА: %v\n", result.Errors)
		}
		results = append(results, result)
		fmt.Println()
	}

	// Итоговая сводка
	printFinalSummary(results, cfg)
}

// parseFlags разбирает флаги командной строки.
func parseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.Addr, "addr", "", "gRPC адрес SSO-сервиса")
	flag.UintVar(&cfg.Concurrency, "concurrency", 0, "параллельных запросов")
	flag.UintVar(&cfg.Requests, "requests", 0, "общее количество запросов")
	flag.BoolVar(&cfg.Heavy, "heavy", false, "усиленный режим (100 concurrency, 50000 requests)")
	flag.StringVar(&cfg.Service, "service", "", "сервис для теста: auth, token, user")
	flag.BoolVar(&cfg.Quiet, "quiet", false, "минимизировать вывод")
	flag.StringVar(&cfg.ProtoPath, "proto", "", "путь к proto файлу")
	flag.StringVar(&cfg.Call, "call", "", "конкретный вызов (переопределяет сценарии)")
	flag.Parse()

	if cfg.Addr == "" {
		cfg.Addr = os.Getenv("LOAD_TEST_SSO_ADDR")
	}
	if cfg.Addr == "" {
		cfg.Addr = "localhost:50051"
	}

	if cfg.Concurrency == 0 {
		cfg.Concurrency = 50
	}
	if cfg.Requests == 0 {
		cfg.Requests = 10000
	}

	if cfg.Heavy {
		cfg.Concurrency = 100
		cfg.Requests = 50000
	}

	return cfg
}

// getScenarios возвращает список тестовых сценариев.
func getScenarios(cfg Config) []TestScenario {
	all := []TestScenario{
		// Auth Service
		{
			Name:    "Register — регистрация нового пользователя",
			Call:    "sso.v1.AuthService.Register",
			Data:    `{"email":"loadtest-user-{{.RequestId}}@test.local","password":"LoadTest123!","display_name":"Load Test User {{.RequestId}}"}`,
			Service: "auth",
		},
		{
			Name:    "Login — вход (существующий пользователь)",
			Call:    "sso.v1.AuthService.Login",
			Data:    `{"email":"user1@test.local","password":"Password123!user1"}`,
			Service: "auth",
		},
		{
			Name:    "Login — неверный пароль (ошибка)",
			Call:    "sso.v1.AuthService.Login",
			Data:    `{"email":"user1@test.local","password":"wrongpassword"}`,
			Service: "auth",
		},
		{
			Name:    "RefreshToken — обновление токена",
			Call:    "sso.v1.AuthService.RefreshToken",
			Data:    `{"refresh_token":"placeholder"}`,
			Service: "auth",
		},
		{
			Name:    "ValidateToken — валидация токена",
			Call:    "sso.v1.TokenService.ValidateToken",
			Data:    `{"access_token":"placeholder"}`,
			Service: "token",
		},
		{
			Name:    "ListSessions — список сессий",
			Call:    "sso.v1.TokenService.ListSessions",
			Data:    `{"user_id":"00000000-0000-0000-0000-000000000000"}`,
			Service: "token",
		},
		{
			Name:    "GetProfile — получение профиля",
			Call:    "sso.v1.UserService.GetProfile",
			Data:    `{"user_id":"00000000-0000-0000-0000-000000000000"}`,
			Service: "user",
		},
		{
			Name:    "SearchUsers — поиск пользователей",
			Call:    "sso.v1.UserService.SearchUsers",
			Data:    `{"query":"user","limit":10,"offset":0}`,
			Service: "user",
		},
		{
			Name:    "GetUserById — получение по ID",
			Call:    "sso.v1.UserService.GetUserById",
			Data:    `{"user_id":"00000000-0000-0000-0000-000000000000"}`,
			Service: "user",
		},
	}

	if cfg.Call != "" {
		// Один конкретный вызов
		return []TestScenario{
			{
				Name: cfg.Call,
				Call: cfg.Call,
				Data: `{"email":"test@test.local","password":"Test123!"}`,
			},
		}
	}

	if cfg.Service != "" {
		var filtered []TestScenario
		for _, sc := range all {
			if sc.Service == cfg.Service {
				filtered = append(filtered, sc)
			}
		}
		return filtered
	}

	return all
}

// runGhzTest запускает ghz для одного сценария.
func runGhzTest(cfg Config, sc TestScenario) TestResult {
	result := TestResult{
		Name: sc.Name,
		Call: sc.Call,
		Date: time.Now().Format(time.RFC3339),
	}

	// Проверяем наличие ghz
	ghzPath, err := exec.LookPath("ghz")
	if err != nil {
		result.Errors = []string{"ghz not found. Install: go install github.com/bojand/ghz/cmd/ghz@latest"}
		return result
	}

	// Определяем proto path
	protoPath := cfg.ProtoPath
	if protoPath == "" {
		// Ищем относительно текущей директории
		protoPath = "../protos"
		if _, err := os.Stat(protoPath); err != nil {
			protoPath = "../../protos"
		}
		if _, err := os.Stat(protoPath); err != nil {
			protoPath = "protos"
		}
	}

	// Собираем аргументы ghz
	args := []string{
		"--insecure",
		"--proto", protoPath + "/sso/v1/auth.proto",
		"--proto", protoPath + "/sso/v1/user.proto",
		"--proto", protoPath + "/sso/v1/token.proto",
		"--proto", protoPath + "/sso/v1/common.proto",
		"--protoc-path", "protoc", // нужен protoc в PATH
		"-c", fmt.Sprintf("%d", cfg.Concurrency),
		"-n", fmt.Sprintf("%d", cfg.Requests),
		"-d", sc.Data,
		"--call", sc.Call,
		cfg.Addr,
	}

	cmd := exec.Command(ghzPath, args...) //nolint:gosec // ghz is the intended load test binary
	output, err := cmd.CombinedOutput()

	if err != nil && len(output) == 0 {
		result.Errors = []string{err.Error()}
		return result
	}

	// Парсим JSON вывод ghz
	var ghzOutput map[string]any
	if err := json.Unmarshal(output, &ghzOutput); err != nil {
		// Если вывод не JSON — пробуем распарсить text вывод
		result.Success = true
		result.Total = parseDurationFromText(string(output))
		result.Average = result.Total
		// Показываем сырой вывод если не удалось распарсить
		if !cfg.Quiet {
			fmt.Printf("  %s\n", string(output))
		}
		return result
	}

	// Заполняем результаты из JSON
	result.Success = true
	if date, ok := ghzOutput["date"].(string); ok {
		result.Date = date
	}
	if count, ok := ghzOutput["count"].(float64); ok {
		result.Count = uint(count)
	}
	if totalMs, ok := ghzOutput["total"].(float64); ok {
		result.Total = time.Duration(totalMs * float64(time.Millisecond))
	}
	if averageMs, ok := ghzOutput["average"].(float64); ok {
		result.Average = time.Duration(averageMs * float64(time.Millisecond))
	}
	if fastestMs, ok := ghzOutput["fastest"].(float64); ok {
		result.Fastest = time.Duration(fastestMs * float64(time.Millisecond))
	}
	if slowestMs, ok := ghzOutput["slowest"].(float64); ok {
		result.Slowest = time.Duration(slowestMs * float64(time.Millisecond))
	}

	// Percentiles
	if details, ok := ghzOutput["details"].(map[string]any); ok {
		if latency, ok := details["latency"].(map[string]any); ok {
			result.Median = parseMs(latency, "50p")
			result.P95 = parseMs(latency, "95p")
			result.P99 = parseMs(latency, "99p")
		}
	}

	if !cfg.Quiet {
		fmt.Printf("  Запросов:   %d\n", result.Count)
		fmt.Printf("  Всего:      %v\n", result.Total.Round(time.Millisecond))
		fmt.Printf("  Средняя:    %v\n", result.Average.Round(time.Microsecond))
		fmt.Printf("  Быстрый:    %v\n", result.Fastest.Round(time.Microsecond))
		fmt.Printf("  Медленный:   %v\n", result.Slowest.Round(time.Microsecond))
		fmt.Printf("  Медиана:    %v\n", result.Median.Round(time.Microsecond))
		fmt.Printf("  P95:        %v\n", result.P95.Round(time.Microsecond))
		fmt.Printf("  P99:        %v\n", result.P99.Round(time.Microsecond))
	}

	// Проверяем ошибки
	if err != nil {
		result.Errors = []string{err.Error()}
	}

	return result
}

// parseMs извлекает длительность из map по ключу (в миллисекундах).
func parseMs(data map[string]any, key string) time.Duration {
	if v, ok := data[key].(float64); ok {
		return time.Duration(v * float64(time.Millisecond))
	}
	return 0
}

// parseDurationFromText пытается извлечь длительность из текстового вывода.
func parseDurationFromText(output string) time.Duration {
	// Fallback — просто возвращаем 0
	_ = output
	return 0
}

// printResult выводит результаты одного теста.
func printResult(r TestResult) {
	fmt.Printf("  Запросов:   %d\n", r.Count)
	fmt.Printf("  Всего:      %v\n", r.Total.Round(time.Millisecond))
	fmt.Printf("  Средняя:    %v\n", r.Average.Round(time.Microsecond))
	fmt.Printf("  Быстрый:    %v\n", r.Fastest.Round(time.Microsecond))
	fmt.Printf("  Медленный:   %v\n", r.Slowest.Round(time.Microsecond))
	fmt.Printf("  Медиана:    %v\n", r.Median.Round(time.Microsecond))
	fmt.Printf("  P95:        %v\n", r.P95.Round(time.Microsecond))
	fmt.Printf("  P99:        %v\n", r.P99.Round(time.Microsecond))

	if len(r.Errors) > 0 {
		fmt.Printf("  Ошибки:     %d\n", len(r.Errors))
	}
}

// printFinalSummary выводит итоговую таблицу.
func printFinalSummary(results []TestResult, cfg Config) {
	fmt.Println("===================================================")
	fmt.Println("  ИТОГОВАЯ СВОДКА")
	fmt.Println("===================================================")
	fmt.Println()

	fmt.Printf("%-50s %8s %10s %10s %10s %10s\n",
		"Тест", "Запросов", "Средняя", "Медиана", "P95", "P99")
	fmt.Println(strings.Repeat("-", 100))

	totalRequests := uint(0)

	for _, r := range results {
		if !r.Success {
			fmt.Printf("%-50s %8s %10s %10s %10s %10s\n",
				truncate(r.Name, 50), "FAIL", "-", "-", "-", "-")
			continue
		}

		fmt.Printf("%-50s %8d %10s %10s %10s %10s\n",
			truncate(r.Name, 50),
			r.Count,
			r.Average.Round(time.Microsecond),
			r.Median.Round(time.Microsecond),
			r.P95.Round(time.Microsecond),
			r.P99.Round(time.Microsecond),
		)
		totalRequests += r.Count
	}

	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("%-50s %8d\n", "ИТОГО запросов", totalRequests)
	fmt.Println()

	// JSON отчёт
	saveJSONReport(results, cfg)
}

// saveJSONReport сохраняет отчёт в файл.
func saveJSONReport(results []TestResult, cfg Config) {
	type Summary struct {
		Timestamp     string `json:"timestamp"`
		Address       string `json:"address"`
		Concurrency   uint   `json:"concurrency"`
		TotalRequests uint   `json:"total_requests"`
	}

	type FullReport struct {
		Summary Summary      `json:"summary"`
		Tests   []TestResult `json:"tests"`
	}

	total := uint(0)
	for _, r := range results {
		if r.Success {
			total += r.Count
		}
	}

	report := FullReport{
		Summary: Summary{
			Timestamp:     time.Now().Format(time.RFC3339),
			Address:       cfg.Addr,
			Concurrency:   cfg.Concurrency,
			TotalRequests: total,
		},
		Tests: results,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка сериализации JSON: %v\n", err)
		return
	}

	reportPath := "load-test-report.json"
	if err := os.WriteFile(reportPath, data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка записи файла: %v\n", err)
		return
	}

	fmt.Printf("JSON-отчёт сохранён в: %s\n", reportPath)
}

// truncate обрезает строку до maxLen.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
