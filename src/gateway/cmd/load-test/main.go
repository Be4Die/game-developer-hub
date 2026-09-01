// Package main provides a lightweight HTTP load-test tool for API Gateway.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	target := flag.String("url", "http://localhost:8080/health", "HTTP target URL")
	concurrency := flag.Int("c", 30, "Number of concurrent workers")
	duration := flag.Duration("d", 10*time.Second, "Test duration")
	flag.Parse()

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	}

	fmt.Printf("Starting HTTP load test against %s (concurrency=%d, duration=%v)...\n", *target, *concurrency, *duration)

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	var (
		totalRequests uint64
		successCount  uint64
		errorCount    uint64
		wg            sync.WaitGroup
	)

	start := time.Now()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				req, err := http.NewRequestWithContext(ctx, http.MethodGet, *target, nil)
				if err != nil {
					return
				}

				atomic.AddUint64(&totalRequests, 1)
				resp, err := client.Do(req)
				if err != nil {
					atomic.AddUint64(&errorCount, 1)
				} else {
					_ = resp.Body.Close()
					if resp.StatusCode >= 200 && resp.StatusCode < 400 {
						atomic.AddUint64(&successCount, 1)
					} else {
						atomic.AddUint64(&errorCount, 1)
					}
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	rps := float64(totalRequests) / elapsed.Seconds()
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Successful:     %d\n", successCount)
	fmt.Printf("Errors:         %d\n", errorCount)
	fmt.Printf("Elapsed Time:   %v\n", elapsed)
	fmt.Printf("Throughput:     %.2f rps\n", rps)
	fmt.Println("═══════════════════════════════════════════════════════")
}
