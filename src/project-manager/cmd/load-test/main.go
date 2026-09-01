// Package main provides a lightweight load-test tool for Project Manager.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	target := flag.String("target", "localhost:50052", "gRPC target address")
	concurrency := flag.Int("c", 20, "Number of concurrent workers")
	duration := flag.Duration("d", 10*time.Second, "Test duration")
	flag.Parse()

	conn, err := grpc.NewClient(*target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", *target, err)
	}
	defer func() { _ = conn.Close() }()

	client := pb.NewProjectServiceClient(conn)

	fmt.Printf("Starting load test against %s (concurrency=%d, duration=%v)...\n", *target, *concurrency, *duration)

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
		go func(workerID int) {
			defer wg.Done()
			userCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(
				"x-user-id", fmt.Sprintf("load-user-%d", workerID),
				"x-user-role", "developer",
			))

			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				atomic.AddUint64(&totalRequests, 1)
				// Call GetPublished (high-frequency read path)
				_, err := client.GetPublished(userCtx, &pb.ProjectGetPublishedRequest{Id: 1})
				if err != nil {
					atomic.AddUint64(&errorCount, 1)
				} else {
					atomic.AddUint64(&successCount, 1)
				}
			}
		}(i)
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
