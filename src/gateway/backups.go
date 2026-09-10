package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	gwpb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	"google.golang.org/grpc/metadata"
)

const gatewayBackupTicketSecret = "gdh-backup-secret-key-2026"

// validateTicket проверяет срок действия и HMAC-подпись одноразового тикета.
func validateTicket(ticket string, nodeID int64, serviceName, backupID string) bool {
	parts := strings.Split(ticket, ":")
	if len(parts) != 5 {
		return false
	}
	tNodeID, _ := strconv.ParseInt(parts[0], 10, 64)
	tServiceName := parts[1]
	tBackupID := parts[2]
	exp, _ := strconv.ParseInt(parts[3], 10, 64)
	sig := parts[4]

	if tNodeID != nodeID || tServiceName != serviceName || tBackupID != backupID {
		return false
	}
	if time.Now().Unix() > exp {
		return false
	}

	payload := fmt.Sprintf("%d:%s:%s:%d", tNodeID, tServiceName, tBackupID, exp)
	mac := hmac.New(sha256.New, []byte(gatewayBackupTicketSecret))
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expectedSig))
}

// handleBackupDownload обрабатывает скачивание бэкапа по одноразовому тикету или JWT.
func handleBackupDownload(nodeClient gwpb.NodeServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeIDStr := r.PathValue("node_id")
		nodeID, err := strconv.ParseInt(nodeIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid node_id", http.StatusBadRequest)
			return
		}

		serviceName := r.PathValue("service_name")
		backupID := r.PathValue("backup_id")

		if serviceName == "" || backupID == "" {
			http.Error(w, "service_name and backup_id are required", http.StatusBadRequest)
			return
		}

		ticket := r.URL.Query().Get("ticket")
		if ticket == "" && r.Header.Get("Authorization") == "" {
			http.Error(w, "authorization required (ticket or bearer token)", http.StatusUnauthorized)
			return
		}

		// Если передан тикет — валидируем криптографическую подпись
		if ticket != "" {
			if !validateTicket(ticket, nodeID, serviceName, backupID) {
				http.Error(w, "invalid or expired download ticket", http.StatusForbidden)
				return
			}
		}

		ctx := authContext(r.Context(), r)
		if ticket != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-backup-ticket", ticket)
		}

		stream, err := nodeClient.DownloadServiceBackup(ctx, &gwpb.NodeServiceDownloadServiceBackupRequest{
			NodeId:      nodeID,
			ServiceName: serviceName,
			BackupId:    backupID,
		})
		if err != nil {
			log.Printf("[ERROR] DownloadServiceBackup: %v", err)
			http.Error(w, fmt.Sprintf("download failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Определяем имя скачиваемого файла
		fileName := backupID
		if !strings.Contains(fileName, ".") {
			fileName += ".archive"
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("[ERROR] stream error during download: %v", err)
				return
			}
			if data := chunk.GetChunk(); len(data) > 0 {
				if _, writeErr := w.Write(data); writeErr != nil {
					return
				}
			}
		}
	}
}

// handleBackupUpload обрабатывает загрузку кастомного файла бэкапа (multipart/form-data).
func handleBackupUpload(nodeClient gwpb.NodeServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			http.Error(w, "multipart/form-data required", http.StatusBadRequest)
			return
		}

		nodeIDStr := r.PathValue("node_id")
		nodeID, err := strconv.ParseInt(nodeIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid node_id", http.StatusBadRequest)
			return
		}

		serviceName := r.PathValue("service_name")
		if serviceName == "" {
			http.Error(w, "service_name is required", http.StatusBadRequest)
			return
		}

		// Лимит на размер файла бэкапа: 2 ГБ
		r.Body = http.MaxBytesReader(w, r.Body, 2<<30)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, fmt.Sprintf("parse multipart form: %v", err), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing 'file' field", http.StatusBadRequest)
			return
		}
		defer file.Close()

		restoreImmediately := r.FormValue("restore_immediately") == "true"

		ctx := authContext(r.Context(), r)
		stream, err := nodeClient.UploadServiceBackup(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("init upload stream: %v", err), http.StatusInternalServerError)
			return
		}

		// 1. Отправляем метаданные первым сообщением
		if err := stream.Send(&gwpb.NodeServiceUploadBackupChunk{
			Payload: &gwpb.NodeServiceUploadBackupChunk_Metadata{
				Metadata: &gwpb.NodeServiceUploadBackupMetadata{
					NodeId:             nodeID,
					ServiceName:        serviceName,
					FileName:           header.Filename,
					RestoreImmediately: restoreImmediately,
				},
			},
		}); err != nil {
			http.Error(w, fmt.Sprintf("send metadata: %v", err), http.StatusInternalServerError)
			return
		}

		// 2. Стримим чанки файла
		err = streamFileChunks(file, func(chunk []byte) error {
			return stream.Send(&gwpb.NodeServiceUploadBackupChunk{
				Payload: &gwpb.NodeServiceUploadBackupChunk_Chunk{
					Chunk: chunk,
				},
			})
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("stream file chunks: %v", err), http.StatusInternalServerError)
			return
		}

		resp, err := stream.CloseAndRecv()
		if err != nil {
			http.Error(w, fmt.Sprintf("finish upload: %v", err), http.StatusInternalServerError)
			return
		}

		writeProtoJSON(w, resp)
	}
}
