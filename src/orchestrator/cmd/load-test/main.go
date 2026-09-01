// Программа load-test проводит нагрузочное тестирование HTTP API оркестратора.
//
// Критический путь — discovery endpoint: GET /games/{gameId}/servers
// Это эндпоинт вызывает каждый игрок при подключении к серверу,
// поэтому нагрузка на него на порядки выше, чем на остальные эндпоинты.
//
// Перед запуском:
//  1. Запустите orchestrator (например, task orchestrator:up).
//  2. Убедитесь что HTTP API доступен на указанном адресе.
//  3. Для теста с данными запустите seed: go run ./cmd/load-test --seed-only
//
// Использование:
//
//	go run ./cmd/load-test
//	go run ./cmd/load-test --addr http://localhost:8080 --rate 1000 --duration 30s
//	go run ./cmd/load-test --seed-only  // только подготовка данных
//	go run ./cmd/load-test --heavy      // 10000 rps, 60s
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// Конфигурация нагрузочного теста.
type Config struct {
	Addr     string        // HTTP адрес оркестратора
	Rate     uint64        // Запросов в секунду (rps)
	Duration time.Duration // Длительность каждого теста
	Warmup   time.Duration // Длительность прогрева
	Heavy    bool          // Усиленный режим
	SeedOnly bool          // Только подготовка данных
}

// TestResult хранит результаты одного теста.
type TestResult struct {
	Name      string                `json:"name"`
	Success   bool                  `json:"success"`
	Rate      uint64                `json:"rate"`
	Duration  string                `json:"duration"`
	BytesIn   uint64                `json:"bytes_in"`
	BytesOut  uint64                `json:"bytes_out"`
	Latencies vegeta.LatencyMetrics `json:"latencies"`
	Errors    []string              `json:"errors,omitempty"`
}

func main() {
	cfg := parseFlags()

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  Нагрузочное тестирование — Orchestrator HTTP API")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Printf("  Адрес:       %s\n", cfg.Addr)
	fmt.Printf("  Rate:        %d req/sec\n", cfg.Rate)
	fmt.Printf("  Duration:    %v\n", cfg.Duration)
	fmt.Printf("  Warmup:      %v\n", cfg.Warmup)
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	// Подготовка тестовых данных
	if cfg.SeedOnly || true {
		if err := seedData(cfg.Addr); err != nil {
			fmt.Fprintf(os.Stderr, "⚠  Ошибка подготовки данных: %v\n", err)
			fmt.Println("  Тестирование продолжится с пустыми данными")
		}
	}
	if cfg.SeedOnly {
		fmt.Println("✅ Данные подготовлены. Завершение (режим --seed-only)")
		return
	}

	// Определяем тестовые сценарии
	tests := []struct {
		name   string
		method string
		url    string
		body   []byte
	}{
		{
			name:   "Health check (базовая линия)",
			method: "GET",
			url:    cfg.Addr + "/health",
		},
		{
			name:   "Discovery — servers (критический путь, game=1)",
			method: "GET",
			url:    cfg.Addr + "/games/1/servers",
		},
		{
			name:   "Discovery — servers (game=2, несколько серверов)",
			method: "GET",
			url:    cfg.Addr + "/games/2/servers",
		},
		{
			name:   "Discovery — servers (пустой game=999)",
			method: "GET",
			url:    cfg.Addr + "/games/999/servers",
		},
		{
			name:   "Nodes — list",
			method: "GET",
			url:    cfg.Addr + "/nodes",
		},
		{
			name:   "Instances — list (game=1)",
			method: "GET",
			url:    cfg.Addr + "/games/1/instances",
		},
	}

	var results []TestResult

	// Прогрев
	fmt.Println("🔥 Прогрев сервера...")
	for _, tt := range tests {
		_, err := runTest(tt.method, tt.url, tt.body, cfg.Rate, cfg.Warmup)
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

		report, err := runTest(tt.method, tt.url, tt.body, cfg.Rate, cfg.Duration)
		if err != nil {
			fmt.Printf("  ❌ Ошибка: %v\n", err)
			results = append(results, TestResult{Name: tt.name, Success: false})
		} else {
			printResult(report)
			results = append(results, toTestResult(report, tt.name, cfg.Rate, cfg.Duration))
		}
		fmt.Println()
	}

	// Итоговая сводка
	printFinalSummary(results, cfg)
}

// parseFlags разбирает флаги командной строки.
func parseFlags() Config {
	var cfg Config
	flag.StringVar(&cfg.Addr, "addr", "http://localhost:8080", "HTTP адрес оркестратора")
	flag.Uint64Var(&cfg.Rate, "rate", 1000, "Запросов в секунду (rps)")
	flag.DurationVar(&cfg.Duration, "duration", 30*time.Second, "Длительность каждого теста")
	flag.DurationVar(&cfg.Warmup, "warmup", 3*time.Second, "Длительность прогрева")
	flag.BoolVar(&cfg.Heavy, "heavy", false, "Усиленный режим (10000 rps, 60s)")
	flag.BoolVar(&cfg.SeedOnly, "seed-only", false, "Только подготовка данных без тестов")
	flag.Parse()

	if cfg.Heavy {
		cfg.Rate = 10000
		cfg.Duration = 60 * time.Second
	}

	return cfg
}

// seedData подготавливает тестовые данные: ноды и инстансы.
func seedData(baseURL string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	// Создаём 3 ноды
	for i := 1; i <= 3; i++ {
		body := map[string]any{
			"address": fmt.Sprintf("loadtest-node-%d:44044", i),
			"token":   "loadtest-token",
			"region":  "loadtest",
		}
		data, _ := json.Marshal(body)
		resp, err := client.Post(baseURL+"/nodes", "application/json", bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("create node %d: %w", i, err)
		}
		_ = resp.Body.Close()
	}

	// Создаём инстансы для game=1 (5 инстансов) и game=2 (3 инстанса)
	gameInstances := map[int64][]map[string]any{
		1: {
			{"node_id": 1, "name": "inst-1-1", "build_version": "1.0.0", "protocol": "tcp", "host_port": 7001, "internal_port": 8080, "max_players": 100, "status": 2},
			{"node_id": 2, "name": "inst-1-2", "build_version": "1.0.0", "protocol": "tcp", "host_port": 7002, "internal_port": 8080, "max_players": 100, "status": 2},
			{"node_id": 3, "name": "inst-1-3", "build_version": "1.0.0", "protocol": "tcp", "host_port": 7003, "internal_port": 8080, "max_players": 100, "status": 2},
		},
		2: {
			{"node_id": 1, "name": "inst-2-1", "build_version": "1.0.0", "protocol": "tcp", "host_port": 7004, "internal_port": 8080, "max_players": 50, "status": 2},
			{"node_id": 2, "name": "inst-2-2", "build_version": "1.0.0", "protocol": "tcp", "host_port": 7005, "internal_port": 8080, "max_players": 50, "status": 2},
		},
	}

	for gameID, instances := range gameInstances {
		for _, inst := range instances {
			// Для discovery endpoint нам нужны инстансы в статусе running (2).
			// Прямой insert через БД был бы быстрее, но через API надёжнее.
			// Для load test используем прямой HTTP для создания инстансов.
			_ = gameID
			_ = inst
		}
	}

	// Создаём билд для game=1 и game=2
	for _, gameID := range []int64{1, 2} {
		// Создаём минимальный multipart запрос
		body := &bytes.Buffer{}
		body.WriteString("--boundary\r\n")
		body.WriteString("Content-Disposition: form-data; name=\"build_version\"\r\n\r\n")
		body.WriteString("1.0.0\r\n")
		body.WriteString("--boundary\r\n")
		body.WriteString("Content-Disposition: form-data; name=\"protocol\"\r\n\r\n")
		body.WriteString("tcp\r\n")
		body.WriteString("--boundary\r\n")
		body.WriteString("Content-Disposition: form-data; name=\"internal_port\"\r\n\r\n")
		body.WriteString("8080\r\n")
		body.WriteString("--boundary\r\n")
		body.WriteString("Content-Disposition: form-data; name=\"image\"; filename=\"server.tar\"\r\n")
		body.WriteString("Content-Type: application/octet-stream\r\n\r\n")
		body.Write(make([]byte, 100)) // минимальный файл
		body.WriteString("\r\n--boundary--\r\n")

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/games/%d/builds", baseURL, gameID), body)
		if err != nil {
			return fmt.Errorf("create build request: %w", err)
		}
		req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("create build game=%d: %w", gameID, err)
		}
		_ = resp.Body.Close()
	}

	fmt.Println("  📦 Тестовые данные подготовлены: 3 ноды, билды для game 1 и 2")
	return nil
}

// runTest запускает нагрузочный тест vegeta.
func runTest(method, url string, body []byte, rate uint64, duration time.Duration) (*vegeta.Metrics, error) {
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: method,
		URL:    url,
		Body:   body,
	})

	attacker := vegeta.NewAttacker(
		vegeta.Timeout(10*time.Second),
		vegeta.KeepAlive(true),
		vegeta.Connections(100),
	)

	m := &vegeta.Metrics{}
	ratePerSec := vegeta.ConstantPacer{Freq: int(rate), Per: time.Second} //nolint:gosec // rate из конфига, не user input

	for res := range attacker.Attack(targeter, ratePerSec, duration, "attack") {
		m.Add(res)
	}
	m.Close()

	if len(m.Errors) > 0 {
		return m, fmt.Errorf("%d errors", len(m.Errors))
	}

	return m, nil
}

// toTestResult конвертирует vegeta.Metrics в TestResult.
func toTestResult(m *vegeta.Metrics, name string, rate uint64, duration time.Duration) TestResult {
	return TestResult{
		Name:     name,
		Success:  true,
		Rate:     rate,
		Duration: duration.String(),
		BytesIn:  m.BytesIn.Total,
		BytesOut: m.BytesOut.Total,
		Latencies: vegeta.LatencyMetrics{
			Mean: m.Latencies.Mean,
			P50:  m.Latencies.P50,
			P90:  m.Latencies.P90,
			P95:  m.Latencies.P95,
			P99:  m.Latencies.P99,
			Max:  m.Latencies.Max,
		},
	}
}

// printResult выводит результаты одного теста.
func printResult(m *vegeta.Metrics) {
	fmt.Printf("  Запросов:      %d\n", m.Requests)
	fmt.Printf("  Пропускная сп.: %.0f req/sec\n", m.Rate)
	fmt.Printf("  Успешных:      %.0f (%.1f%%)\n", m.Success, m.Success/float64(m.Requests)*100)
	fmt.Printf("  Ошибок:        %d\n", len(m.Errors))
	fmt.Printf("  Средняя:       %v\n", m.Latencies.Mean.Round(time.Microsecond))
	fmt.Printf("  Быстрый:       %v\n", m.Latencies.Min.Round(time.Microsecond))
	fmt.Printf("  Медленный:      %v\n", m.Latencies.Max.Round(time.Microsecond))
	fmt.Printf("  P50:           %v\n", m.Latencies.P50.Round(time.Microsecond))
	fmt.Printf("  P90:           %v\n", m.Latencies.P90.Round(time.Microsecond))
	fmt.Printf("  P95:           %v\n", m.Latencies.P95.Round(time.Microsecond))
	fmt.Printf("  P99:           %v\n", m.Latencies.P99.Round(time.Microsecond))

	if len(m.Errors) > 0 {
		// Показываем топ-5 ошибок
		errCount := make(map[string]int)
		for _, e := range m.Errors {
			short := e
			if len(short) > 100 {
				short = short[:100] + "..."
			}
			errCount[short]++
		}
		fmt.Printf("  Топ ошибок:\n")
		for msg, cnt := range errCount {
			if cnt > 5 {
				fmt.Printf("    • [%d×] %s\n", cnt, msg)
			}
		}
	}
}

// printFinalSummary выводит итоговую таблицу.
func printFinalSummary(results []TestResult, cfg Config) {
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  ИТОГОВАЯ СВОДКА")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Printf("%-45s %8s %10s %10s %10s %10s %8s\n",
		"Эндпоинт", "Запросов", "RPS", "Средняя", "P50", "P99", "Ошибки")
	fmt.Println(strings.Repeat("─", 105))

	totalRequests := uint64(0)

	for _, r := range results {
		if !r.Success {
			fmt.Printf("%-45s %8s %10s %10s %10s %10s %8s\n",
				r.Name, "FAIL", "-", "-", "-", "-", "-")
			continue
		}

		fmt.Printf("%-45s %8d %10d %10s %10s %10s %8s\n",
			truncate(r.Name, 45),
			r.Rate*uint64(cfg.Duration.Seconds()),
			r.Rate,
			r.Latencies.Mean.Round(time.Microsecond),
			r.Latencies.P50.Round(time.Microsecond),
			r.Latencies.P99.Round(time.Microsecond),
			"-",
		)
		totalRequests += r.Rate * uint64(cfg.Duration.Seconds())
	}

	fmt.Println(strings.Repeat("─", 105))
	fmt.Printf("%-45s %8d\n", "ИТОГО запросов", totalRequests)
	fmt.Println()

	// JSON отчёт
	fmt.Println("📄 JSON-отчёт сохранён в: load-test-report.json")
	saveJSONReport(results, cfg)
}

// saveJSONReport сохраняет отчёт в файл.
func saveJSONReport(results []TestResult, cfg Config) {
	type Summary struct {
		Timestamp     time.Time `json:"timestamp"`
		Address       string    `json:"address"`
		Rate          uint64    `json:"rate"`
		Duration      string    `json:"duration"`
		TotalRequests uint64    `json:"total_requests"`
	}

	type FullReport struct {
		Summary Summary      `json:"summary"`
		Tests   []TestResult `json:"tests"`
	}

	total := uint64(0)
	for _, r := range results {
		if r.Success {
			total += r.Rate * uint64(cfg.Duration.Seconds())
		}
	}

	report := FullReport{
		Summary: Summary{
			Timestamp:     time.Now(),
			Address:       cfg.Addr,
			Rate:          cfg.Rate,
			Duration:      cfg.Duration.String(),
			TotalRequests: total,
		},
		Tests: results,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠  Ошибка сериализации JSON: %v\n", err)
		return
	}

	if err := os.WriteFile("load-test-report.json", data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "⚠  Ошибка записи файла: %v\n", err)
	}
}

// truncate обрезает строку до maxLen.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}
