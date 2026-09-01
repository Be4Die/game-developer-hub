package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"

	gwpb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

// handleInstanceLogsStream обрабатывает GET /api/v1/games/{game_id}/instances/{instance_id}/logs (SSE стриминг логов).
func handleInstanceLogsStream(client gwpb.InstanceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameID, err := strconv.ParseInt(r.PathValue("game_id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid game_id", http.StatusBadRequest)
			return
		}

		instanceID, err := strconv.ParseInt(r.PathValue("instance_id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid instance_id", http.StatusBadRequest)
			return
		}

		follow := r.URL.Query().Get("follow") == "true"
		tail := int32(100)
		if t, err := strconv.ParseInt(r.URL.Query().Get("tail"), 10, 32); err == nil && t > 0 {
			tail = int32(t)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ctx := r.Context()
		token := r.Header.Get("Authorization")
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		if token != "" {
			if !strings.HasPrefix(token, "Bearer ") {
				token = "Bearer " + token
			}
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
		}
		if apiKey := r.Header.Get("X-Api-Key"); apiKey != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-api-key", apiKey)
		}

		stream, err := client.StreamLogs(ctx, &gwpb.InstanceServiceStreamLogsRequest{
			GameId:     gameID,
			InstanceId: instanceID,
			Follow:     follow,
			Tail:       tail,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("create logs stream: %v", err), http.StatusInternalServerError)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			resp, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Printf("logs stream error: %v", err)
				}
				return
			}

			fmt.Fprintf(w, "event: log\ndata: %s\n\n", formatLogEvent(resp.Entry))
			flusher.Flush()
		}
	}
}

func formatLogEvent(entry *gwpb.LogEntry) string {
	if entry == nil {
		return "{}"
	}
	return fmt.Sprintf(`{"timestamp":"%s","source":"%s","message":%s}`,
		entry.Timestamp.AsTime().Format(time.RFC3339),
		entry.Source.String(),
		strconv.Quote(entry.Message),
	)
}
