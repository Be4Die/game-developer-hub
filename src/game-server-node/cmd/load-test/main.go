// Программа load-test проводит нагрузочное тестирование DiscoveryService.
//
// Перед запуском:
//  1. Запустите game-server-node (task node:up или вручную).
//  2. Убедитесь что gRPC доступен на указанном адресе.
//
// Использование:
//
//	go run ./cmd/load-test --addr localhost:44044
//	go run ./cmd/load-test --addr localhost:44044 --concurrency 100 --requests 5000
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bojand/ghz/runner"
)

// Конфигурация нагрузочного теста.
type Config struct {
	Addr        string // gRPC адрес: host:port
	Concurrency uint   // Параллельных запросов
	Requests    uint   // Всего запросов на метод
	Warmup      uint   // Прогревочных запросов на метод
	ProtoDir    string // Путь к директории с proto файлами
	Timeout     time.Duration
}

// TestResult хранит агрегированные результаты теста одного метода.
type TestResult struct {
	Method  string
	Success bool
	Count   uint64
	Total   time.Duration
	Average time.Duration
	Fastest time.Duration
	Slowest time.Duration
	Rps     float64
	P50     time.Duration
	P90     time.Duration
	P95     time.Duration
	P99     time.Duration
	Errors  map[string]int
	Report  *runner.Report
}

func main() {
	cfg := parseFlags()

	protoFile := filepath.Join(cfg.ProtoDir, "game_server_node", "v1", "discovery.proto")
	importPaths := []string{cfg.ProtoDir}

	// Проверяем доступность proto файла
	if _, err := os.Stat(protoFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "proto file not found: %s\n", protoFile)
		os.Exit(1)
	}

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  Нагрузочное тестирование — DiscoveryService")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Printf("  Адрес:       %s\n", cfg.Addr)
	fmt.Printf("  Concurrency: %d\n", cfg.Concurrency)
	fmt.Printf("  Requests:    %d (на метод)\n", cfg.Requests)
	fmt.Printf("  Warmup:      %d (на метод)\n", cfg.Warmup)
	fmt.Printf("  Timeout:     %v\n", cfg.Timeout)
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	// Определяем тесты для всех методов DiscoveryService
	tests := []struct {
		name     string
		call     string
		dataJSON string
	}{
		{
			name:     "GetNodeInfo",
			call:     "game_server_node.v1.DiscoveryService.GetNodeInfo",
			dataJSON: `{}`,
		},
		{
			name:     "Heartbeat",
			call:     "game_server_node.v1.DiscoveryService.Heartbeat",
			dataJSON: `{}`,
		},
		{
			name:     "ListInstances",
			call:     "game_server_node.v1.DiscoveryService.ListInstances",
			dataJSON: `{}`,
		},
		{
			name:     "GetInstance (существующий)",
			call:     "game_server_node.v1.DiscoveryService.GetInstance",
			dataJSON: `{"instance_id": 1}`,
		},
		{
			name:     "GetInstance (не существующий)",
			call:     "game_server_node.v1.DiscoveryService.GetInstance",
			dataJSON: `{"instance_id": 999999}`,
		},
		{
			name:     "ListInstancesByGame",
			call:     "game_server_node.v1.DiscoveryService.ListInstancesByGame",
			dataJSON: `{"game_id": 1}`,
		},
		{
			name:     "GetInstanceUsage",
			call:     "game_server_node.v1.DiscoveryService.GetInstanceUsage",
			dataJSON: `{"instance_id": 1}`,
		},
		{
			name:     "GetInstanceUsage (не существующий)",
			call:     "game_server_node.v1.DiscoveryService.GetInstanceUsage",
			dataJSON: `{"instance_id": 999999}`,
		},
	}

	var results []TestResult

	// Сначала прогрев
	fmt.Println("🔥 Прогрев сервера...")
	for _, tt := range tests {
		_, err := runner.Run(
			tt.call,
			cfg.Addr,
			runner.WithProtoFile(protoFile, importPaths),
			runner.WithDataFromJSON(tt.dataJSON),
			runner.WithInsecure(true),
			runner.WithConcurrency(cfg.Warmup),
			runner.WithTotalRequests(cfg.Warmup),
			runner.WithTimeout(cfg.Timeout),
		)
		if err != nil {
			fmt.Printf("  ⚠  Прогрев %s: %v\n", tt.name, err)
		}
	}
	fmt.Println("  Прогрев завершён")
	fmt.Println()

	// Основной прогон
	for i, tt := range tests {
		fmt.Printf("[%d/%d] Тест: %s\n", i+1, len(tests), tt.name)
		fmt.Println(strings.Repeat("─", 55))

		report, err := runner.Run(
			tt.call,
			cfg.Addr,
			runner.WithProtoFile(protoFile, importPaths),
			runner.WithDataFromJSON(tt.dataJSON),
			runner.WithInsecure(true),
			runner.WithConcurrency(cfg.Concurrency),
			runner.WithTotalRequests(cfg.Requests),
			runner.WithTimeout(cfg.Timeout),
			runner.WithName(tt.name),
		)

		result := TestResult{
			Method: tt.name,
			Report: report,
		}

		if err != nil {
			result.Success = false
			fmt.Printf("  ❌ Ошибка: %v\n", err)
		} else {
			result.Success = true
			result.Count = report.Count
			result.Total = report.Total
			result.Average = report.Average
			result.Fastest = report.Fastest
			result.Slowest = report.Slowest
			result.Rps = report.Rps
			result.Errors = report.ErrorDist

			for _, ld := range report.LatencyDistribution {
				switch ld.Percentage {
				case 50:
					result.P50 = ld.Latency
				case 90:
					result.P90 = ld.Latency
				case 95:
					result.P95 = ld.Latency
				case 99:
					result.P99 = ld.Latency
				}
			}

			printMethodSummary(result)
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

	flag.StringVar(&cfg.Addr, "addr", "localhost:44044", "gRPC сервер адрес (host:port)")
	flag.UintVar(&cfg.Concurrency, "concurrency", 50, "Количество параллельных запросов (-c)")
	flag.UintVar(&cfg.Requests, "requests", 10000, "Общее количество запросов на метод (-n)")
	flag.UintVar(&cfg.Warmup, "warmup", 100, "Прогревочных запросов на метод")
	flag.DurationVar(&cfg.Timeout, "timeout", 30*time.Second, "Таймаут одного запроса")

	// Путь к proto файлам — ищем относительно текущей директории или используем окружение
	defaultProtoDir := detectProtoDir()
	flag.StringVar(&cfg.ProtoDir, "proto-dir", defaultProtoDir, "Путь к корневой директории proto файлов")

	flag.Parse()

	return cfg
}

// detectProtoDir пытается найти директорию с proto файлами.
func detectProtoDir() string {
	// Варианты поиска
	candidates := []string{
		// Относительно game-server-node/
		"../protos",
		// Относительно workspace
		"../../protos",
		// Абсолютный путь (Windows)
		filepath.Join(os.Getenv("USERPROFILE"), "Documents", "game-developer-hub", "protos"),
	}

	for _, path := range candidates {
		abs, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		testPath := filepath.Join(abs, "game_server_node", "v1", "discovery.proto")
		if _, err := os.Stat(filepath.Clean(testPath)); err == nil {
			return abs
		}
	}

	// Фоллбэк — текущая директория
	return "."
}

// printMethodSummary выводит результаты одного теста.
func printMethodSummary(r TestResult) {
	if !r.Success {
		return
	}

	fmt.Printf("  Запросов:      %d\n", r.Count)
	fmt.Printf("  Пропускная сп.: %.0f req/sec\n", r.Rps)
	fmt.Printf("  Средняя:       %v\n", r.Average.Round(time.Microsecond))
	fmt.Printf("  Быстрый:       %v\n", r.Fastest.Round(time.Microsecond))
	fmt.Printf("  Медленный:      %v\n", r.Slowest.Round(time.Microsecond))
	fmt.Printf("  P50:           %v\n", r.P50.Round(time.Microsecond))
	fmt.Printf("  P90:           %v\n", r.P90.Round(time.Microsecond))
	fmt.Printf("  P95:           %v\n", r.P95.Round(time.Microsecond))
	fmt.Printf("  P99:           %v\n", r.P99.Round(time.Microsecond))

	if len(r.Errors) > 0 {
		fmt.Printf("  Ошибки:        %d total\n", sumMap(r.Errors))
		for msg, count := range r.Errors {
			// Обрезаем длинные сообщения
			short := msg
			if len(short) > 100 {
				short = short[:100] + "..."
			}
			fmt.Printf("    • [%d×] %s\n", count, short)
		}
	} else {
		fmt.Println("  Ошибки:        0")
	}
}

// printFinalSummary выводит итоговую таблицу результатов.
func printFinalSummary(results []TestResult, cfg Config) {
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  ИТОГОВАЯ СВОДКА")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	// Таблица
	fmt.Printf("%-35s %8s %10s %10s %10s %10s %8s\n",
		"Метод", "Запросов", "RPS", "Средняя", "P50", "P99", "Ошибки")
	fmt.Println(strings.Repeat("─", 95))

	totalRequests := uint64(0)
	totalErrors := 0
	allSuccess := true

	for _, r := range results {
		if !r.Success {
			fmt.Printf("%-35s %8s %10s %10s %10s %10s %8s\n",
				r.Method, "FAIL", "-", "-", "-", "-", "-")
			allSuccess = false
			continue
		}

		totalRequests += r.Count
		errorCount := 0
		for _, c := range r.Errors {
			errorCount += c
		}
		totalErrors += errorCount

		fmt.Printf("%-35s %8d %10.0f %10s %10s %10s %8d\n",
			truncate(r.Method, 35),
			r.Count,
			r.Rps,
			r.Average.Round(time.Microsecond),
			r.P50.Round(time.Microsecond),
			r.P99.Round(time.Microsecond),
			errorCount,
		)
	}

	fmt.Println(strings.Repeat("─", 95))
	fmt.Printf("%-35s %8d %10s %10s %10s %10s %8d\n",
		"ИТОГО", totalRequests, "", "", "", "", totalErrors)
	fmt.Println()

	// JSON отчёт
	fmt.Println("📄 JSON-отчёт сохранён в: load-test-report.json")
	saveJSONReport(results, cfg)

	if allSuccess && totalErrors == 0 {
		fmt.Println("✅ Все тесты прошли без ошибок")
	} else {
		fmt.Printf("⚠️  Обнаружено %d ошибок\n", totalErrors)
	}
}

// saveJSONReport сохраняет детальный отчёт в файл.
func saveJSONReport(results []TestResult, cfg Config) {
	type ReportSummary struct {
		Timestamp     time.Time     `json:"timestamp"`
		Address       string        `json:"address"`
		Concurrency   uint          `json:"concurrency"`
		Requests      uint          `json:"requests_per_method"`
		Warmup        uint          `json:"warmup"`
		Timeout       time.Duration `json:"timeout"`
		TotalRequests uint64        `json:"total_requests"`
		TotalErrors   int           `json:"total_errors"`
	}

	type FullReport struct {
		Summary ReportSummary  `json:"summary"`
		Methods map[string]any `json:"methods"`
	}

	report := FullReport{
		Summary: ReportSummary{
			Timestamp:   time.Now(),
			Address:     cfg.Addr,
			Concurrency: cfg.Concurrency,
			Requests:    cfg.Requests,
			Warmup:      cfg.Warmup,
			Timeout:     cfg.Timeout,
		},
		Methods: make(map[string]any),
	}

	for _, r := range results {
		if !r.Success || r.Report == nil {
			continue
		}

		report.Summary.TotalRequests += r.Count
		for _, c := range r.Errors {
			report.Summary.TotalErrors += c
		}

		report.Methods[r.Method] = map[string]any{
			"count":   r.Count,
			"rps":     r.Rps,
			"average": r.Average.String(),
			"fastest": r.Fastest.String(),
			"slowest": r.Slowest.String(),
			"p50":     r.P50.String(),
			"p90":     r.P90.String(),
			"p95":     r.P95.String(),
			"p99":     r.P99.String(),
			"errors":  r.Errors,
		}
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Ошибка сериализации JSON: %v\n", err)
		return
	}

	if err := os.WriteFile("load-test-report.json", data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Ошибка записи файла: %v\n", err)
	}
}

// truncate обрезает строку до maxLen символов.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}

// sumMap суммирует значения в map.
func sumMap(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}
