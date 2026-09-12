package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"google.golang.org/grpc/metadata"

	modpb "github.com/Be4Die/game-developer-hub/protos/moderation/v1"
	projpb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
)

const (
	maxImageSizeBytes = 15 * 1024 * 1024 // 15 МБ
	maxVideoSizeBytes = 50 * 1024 * 1024 // 50 МБ
)

func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // RFC 4122 v4
	b[8] = (b[8] & 0x3f) | 0x80 // variant
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// decodeJWTSegment декодирует сегмент JWT с поддержкой raw URL, padded URL и стандартного Base64.
func decodeJWTSegment(seg string) ([]byte, error) {
	seg = strings.TrimSpace(seg)
	if b, err := base64.RawURLEncoding.DecodeString(seg); err == nil {
		return b, nil
	}
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}
	if b, err := base64.URLEncoding.DecodeString(seg); err == nil {
		return b, nil
	}
	return base64.StdEncoding.DecodeString(seg)
}

// extractAuthUser извлекает userID, userRole и rawToken из Authorization заголовка, query параметров или cookies.
func extractAuthUser(r *http.Request) (userID, userRole, rawToken string) {
	if auth := r.Header.Get("Authorization"); auth != "" {
		trimmed := strings.TrimSpace(auth)
		if strings.HasPrefix(strings.ToLower(trimmed), "bearer ") {
			rawToken = strings.TrimSpace(trimmed[7:])
		} else {
			rawToken = trimmed
		}
	}
	if rawToken == "" {
		if tok := r.URL.Query().Get("token"); tok != "" {
			rawToken = strings.TrimSpace(tok)
		} else if tok := r.URL.Query().Get("auth"); tok != "" {
			rawToken = strings.TrimSpace(tok)
		} else if tok := r.URL.Query().Get("access_token"); tok != "" {
			rawToken = strings.TrimSpace(tok)
		}
	}
	if rawToken == "" {
		if cookie, err := r.Cookie("gdh_access_token"); err == nil && cookie.Value != "" {
			rawToken = strings.TrimSpace(cookie.Value)
		} else if cookie, err := r.Cookie("gdh_session"); err == nil && cookie.Value != "" {
			rawToken = strings.TrimSpace(cookie.Value)
		} else if cookie, err := r.Cookie("access_token"); err == nil && cookie.Value != "" {
			rawToken = strings.TrimSpace(cookie.Value)
		}
	}

	if rawToken == "" {
		return "", "", ""
	}

	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return "", "", rawToken
	}

	payload, err := decodeJWTSegment(parts[1])
	if err != nil {
		return "", "", rawToken
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", "", rawToken
	}

	if sub, ok := claims["sub"].(string); ok {
		userID = sub
	} else if uid, ok := claims["user_id"].(string); ok {
		userID = uid
	} else if id, ok := claims["id"].(string); ok {
		userID = id
	}

	var rawRole any
	if rVal, ok := claims["role"]; ok {
		rawRole = rVal
	} else if rVal, ok := claims["user_role"]; ok {
		rawRole = rVal
	} else if rVal, ok := claims["roles"]; ok {
		rawRole = rVal
	}

	switch v := rawRole.(type) {
	case float64:
		userRole = fmt.Sprintf("%.0f", v)
	case int:
		userRole = strconv.Itoa(v)
	case int64:
		userRole = strconv.FormatInt(v, 10)
	case json.Number:
		userRole = v.String()
	case string:
		userRole = v
	case []any:
		if len(v) > 0 {
			userRole = fmt.Sprintf("%v", v[0])
		}
	}

	return userID, userRole, rawToken
}

// isRoleModeratorOrAdmin определяет, обладает ли роль правами модератора или администратора.
func isRoleModeratorOrAdmin(role string) bool {
	r := strings.TrimSpace(strings.ToLower(role))
	return r == "admin" || r == "moderator" ||
		r == "user_role_admin" || r == "user_role_moderator" ||
		r == "role_admin" || r == "role_moderator" ||
		r == "2" || r == "3"
}

// normalizeRoleString возвращает числовую строку роли ("1" = developer, "2" = moderator, "3" = admin).
func normalizeRoleString(role string) string {
	r := strings.TrimSpace(strings.ToLower(role))
	switch {
	case r == "admin" || r == "user_role_admin" || r == "role_admin" || r == "3":
		return "3"
	case r == "moderator" || r == "user_role_moderator" || r == "role_moderator" || r == "2":
		return "2"
	default:
		return "1"
	}
}

// outgoingAuthContext создает контекст с исходящими gRPC метаданными авторизации (x-user-id, x-user-role, authorization).
func outgoingAuthContext(ctx context.Context, userID, userRole, rawToken string) context.Context {
	md := metadata.MD{}
	if userID != "" {
		md.Set("x-user-id", userID)
	}
	if userRole != "" {
		md.Set("x-user-role", normalizeRoleString(userRole))
	}
	if rawToken != "" {
		md.Set("authorization", "Bearer "+rawToken)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// checkProjectAccess проверяет, имеет ли пользователь право доступа к проекту (модератор, админ или владелец/участник).
func checkProjectAccess(
	ctx context.Context,
	projClient projpb.ProjectServiceClient,
	projectID int64,
	userID, userRole, rawToken string,
) bool {
	if isRoleModeratorOrAdmin(userRole) {
		return true
	}
	if userID == "" || projClient == nil {
		return false
	}

	callCtx := outgoingAuthContext(ctx, userID, userRole, rawToken)

	// 1. Проверяем владельца и участника через Get (в Project Manager метод GetProjectForUser
	// автоматически проверяет, является ли пользователь владельцем или участником с правами)
	projResp, err := projClient.Get(callCtx, &projpb.ProjectGetRequest{Id: projectID})
	if err == nil && projResp.GetProject() != nil {
		return true
	}

	// 2. Резервная проверка через список участников проекта
	membersResp, errMembers := projClient.ListMembers(callCtx, &projpb.ProjectListMembersRequest{ProjectId: projectID})
	if errMembers == nil && membersResp != nil {
		if membersResp.GetOwnerId() == userID {
			return true
		}
		for _, m := range membersResp.GetMembers() {
			if m.GetUserId() == userID {
				return true
			}
		}
	}

	log.Printf("[chat_attachments] access denied for user=%s role=%s on project=%d (getErr: %v, membersErr: %v)",
		userID, userRole, projectID, err, errMembers)
	return false
}

func sendJSONError(w http.ResponseWriter, msg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   msg,
		"message": msg,
	})
}

// handleChatAttachmentUpload обрабатывает POST /api/v1/projects/{project_id}/chat/attachments
func handleChatAttachmentUpload(
	modClient modpb.ModerationServiceClient,
	projClient projpb.ProjectServiceClient,
	basePath string,
	s3Clients ...*s3.Client,
) http.HandlerFunc {
	var s3Client *s3.Client
	if len(s3Clients) > 0 {
		s3Client = s3Clients[0]
	}

	attBucket := os.Getenv("S3_ATTACHMENTS_BUCKET")
	if attBucket == "" {
		attBucket = os.Getenv("S3_MEDIA_BUCKET")
		if attBucket == "" {
			attBucket = "media"
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			sendJSONError(w, "multipart/form-data required", http.StatusBadRequest)
			return
		}

		projectIDStr := r.PathValue("project_id")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil || projectID <= 0 {
			sendJSONError(w, "invalid project_id", http.StatusBadRequest)
			return
		}

		// Авторизация
		userID, userRole, rawToken := extractAuthUser(r)
		if userID == "" {
			sendJSONError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if !checkProjectAccess(r.Context(), projClient, projectID, userID, userRole, rawToken) {
			sendJSONError(w, "forbidden: no access to project chat", http.StatusForbidden)
			return
		}

		// Ограничение размера тела запроса (макс. 55 МБ)
		r.Body = http.MaxBytesReader(w, r.Body, (maxVideoSizeBytes + 5*1024*1024))
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, fmt.Sprintf("file too large or parse error: %v", err), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing 'file' form field", http.StatusBadRequest)
			return
		}
		defer func() { _ = file.Close() }()

		cleanFilename := filepath.Base(filepath.Clean(header.Filename))
		ext := strings.ToLower(filepath.Ext(cleanFilename))

		// Определение и валидация MIME-типа
		detectedMime := header.Header.Get("Content-Type")
		if detectedMime == "" || detectedMime == "application/octet-stream" {
			detectedMime = mime.TypeByExtension(ext)
		}
		detectedMime = strings.ToLower(strings.Split(detectedMime, ";")[0])

		isImage := strings.HasPrefix(detectedMime, "image/") || ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp"
		isVideo := strings.HasPrefix(detectedMime, "video/") || ext == ".mp4" || ext == ".webm"

		if !isImage && !isVideo {
			http.Error(w, "unsupported file format: only images (PNG, JPEG, WebP) and videos (MP4, WebM) are allowed", http.StatusBadRequest)
			return
		}

		if isImage {
			if detectedMime != "image/png" && detectedMime != "image/jpeg" && detectedMime != "image/webp" {
				if ext == ".png" {
					detectedMime = "image/png"
				} else if ext == ".webp" {
					detectedMime = "image/webp"
				} else {
					detectedMime = "image/jpeg"
				}
			}
			if header.Size > maxImageSizeBytes {
				http.Error(w, fmt.Sprintf("image size exceeds limit of 15 MB (size: %.1f MB)", float64(header.Size)/(1024*1024)), http.StatusBadRequest)
				return
			}
		}

		if isVideo {
			if detectedMime != "video/mp4" && detectedMime != "video/webm" {
				if ext == ".webm" {
					detectedMime = "video/webm"
				} else {
					detectedMime = "video/mp4"
				}
			}
			if header.Size > maxVideoSizeBytes {
				http.Error(w, fmt.Sprintf("video size exceeds limit of 50 MB (size: %.1f MB)", float64(header.Size)/(1024*1024)), http.StatusBadRequest)
				return
			}
		}

		attID := newUUID()
		storageSubPath := fmt.Sprintf("projects/%d/chat/%s_%s", projectID, attID, cleanFilename)

		// 1. Сохранение в S3 / SeaweedFS
		uploadedToS3 := false
		if s3Client != nil {
			_, putErr := s3Client.PutObject(r.Context(), &s3.PutObjectInput{
				Bucket:        aws.String(attBucket),
				Key:           aws.String(storageSubPath),
				Body:          file,
				ContentType:   aws.String(detectedMime),
				ContentLength: aws.Int64(header.Size),
			})
			if putErr == nil {
				uploadedToS3 = true
			} else {
				log.Printf("failed to upload attachment to S3: %v, falling back to local disk", putErr)
				// rewind file for local save
				if seeker, ok := file.(io.Seeker); ok {
					_, _ = seeker.Seek(0, io.SeekStart)
				}
			}
		}

		// 2. Фоллбэк на локальный диск
		if !uploadedToS3 {
			localDir := filepath.Join(basePath, "chat_attachments", fmt.Sprintf("%d", projectID))
			if err := os.MkdirAll(localDir, 0750); err != nil {
				log.Printf("cannot create dir in basePath (%s): %v, falling back to temp dir", localDir, err)
				localDir = filepath.Join(os.TempDir(), "chat_attachments", fmt.Sprintf("%d", projectID))
				if err2 := os.MkdirAll(localDir, 0750); err2 != nil {
					http.Error(w, fmt.Sprintf("create attachment dir: %v", err2), http.StatusInternalServerError)
					return
				}
			}
			localFilePath := filepath.Join(localDir, attID+"_"+cleanFilename)
			dst, createErr := os.Create(localFilePath)
			if createErr != nil {
				http.Error(w, fmt.Sprintf("create attachment file: %v", createErr), http.StatusInternalServerError)
				return
			}
			defer func() { _ = dst.Close() }()
			if _, copyErr := io.Copy(dst, file); copyErr != nil {
				http.Error(w, fmt.Sprintf("write attachment file: %v", copyErr), http.StatusInternalServerError)
				return
			}
			storageSubPath = localFilePath
		}

		// 3. Регистрация в Moderation Service
		uploaderRole := modpb.SenderRole_SENDER_ROLE_DEVELOPER
		if isRoleModeratorOrAdmin(userRole) {
			uploaderRole = modpb.SenderRole_SENDER_ROLE_MODERATOR
		}

		callCtx := outgoingAuthContext(r.Context(), userID, userRole, rawToken)
		regResp, regErr := modClient.RegisterAttachment(callCtx, &modpb.RegisterAttachmentRequest{
			Id:           attID,
			ProjectId:    projectID,
			UploaderId:   userID,
			UploaderRole: uploaderRole,
			FileName:     cleanFilename,
			FileSize:     header.Size,
			MimeType:     detectedMime,
			StoragePath:  storageSubPath,
		})
		if regErr != nil {
			log.Printf("failed to register attachment in moderation service: %v", regErr)
			http.Error(w, fmt.Sprintf("register attachment: %v", regErr), http.StatusInternalServerError)
			return
		}

		writeProtoJSON(w, regResp.GetAttachment())
	}
}

// handleChatAttachmentServe обрабатывает GET /api/v1/projects/{project_id}/chat/attachments/{attachment_id} и .../download
func handleChatAttachmentServe(
	modClient modpb.ModerationServiceClient,
	projClient projpb.ProjectServiceClient,
	basePath string,
	s3Clients ...*s3.Client,
) http.HandlerFunc {
	var s3Client *s3.Client
	if len(s3Clients) > 0 {
		s3Client = s3Clients[0]
	}

	attBucket := os.Getenv("S3_ATTACHMENTS_BUCKET")
	if attBucket == "" {
		attBucket = os.Getenv("S3_MEDIA_BUCKET")
		if attBucket == "" {
			attBucket = "media"
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		projectIDStr := r.PathValue("project_id")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil || projectID <= 0 {
			http.NotFound(w, r)
			return
		}

		attID := r.PathValue("attachment_id")
		if attID == "" {
			http.NotFound(w, r)
			return
		}

		// Проверка авторизации и прав
		userID, userRole, rawToken := extractAuthUser(r)
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if !checkProjectAccess(r.Context(), projClient, projectID, userID, userRole, rawToken) {
			http.Error(w, "forbidden: access denied", http.StatusForbidden)
			return
		}

		// Получаем метаданные вложения
		callCtx := outgoingAuthContext(r.Context(), userID, userRole, rawToken)
		attResp, err := modClient.GetAttachment(callCtx, &modpb.GetAttachmentRequest{
			AttachmentId: attID,
			ProjectId:    projectID,
		})
		if err != nil || attResp.GetAttachment() == nil {
			http.NotFound(w, r)
			return
		}

		att := attResp.GetAttachment()
		if att.GetIsPurged() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusGone)
			_, _ = w.Write([]byte(`{"error":"media purged after project publication"}`))
			return
		}

		// Заголовки скачивания
		isDownload := strings.HasSuffix(r.URL.Path, "/download") || r.URL.Query().Get("download") == "1"
		cleanFilename := att.GetFileName()
		if cleanFilename == "" {
			cleanFilename = "attachment"
		}

		if isDownload {
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", cleanFilename))
		} else {
			w.Header().Set("Content-Disposition", "inline")
		}

		contentType := att.GetMimeType()
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", contentType)

		// 1. Попытка отдать из S3 (SeaweedFS)
		if s3Client != nil {
			s3Key := fmt.Sprintf("projects/%d/chat/%s_%s", projectID, attID, cleanFilename)
			getInput := &s3.GetObjectInput{
				Bucket: aws.String(attBucket),
				Key:    aws.String(s3Key),
			}
			if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
				getInput.Range = aws.String(rangeHeader)
			}

			out, s3Err := s3Client.GetObject(r.Context(), getInput)
			if s3Err == nil {
				defer func() { _ = out.Body.Close() }()

				// Обработка Range запросов (206 Partial Content для видеоплеера)
				if out.ContentRange != nil && *out.ContentRange != "" {
					w.Header().Set("Content-Range", *out.ContentRange)
					w.WriteHeader(http.StatusPartialContent)
				} else {
					w.Header().Set("Accept-Ranges", "bytes")
				}

				if out.ContentLength != nil && *out.ContentLength > 0 {
					w.Header().Set("Content-Length", fmt.Sprintf("%d", *out.ContentLength))
				}

				_, _ = io.Copy(w, out.Body)
				return
			}
			// Проверяем ошибку 404 vs ошибка диапазона
			var nsk *s3types.NoSuchKey
			if errors.As(s3Err, &nsk) {
				log.Printf("s3 key not found: %s, checking local disk", s3Key)
			}
		}

		// 2. Фоллбэк на локальный диск
		localDir := filepath.Join(basePath, "chat_attachments", fmt.Sprintf("%d", projectID))
		localFilePath := filepath.Join(localDir, attID+"_"+cleanFilename)
		if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
			tempFilePath := filepath.Join(os.TempDir(), "chat_attachments", fmt.Sprintf("%d", projectID), attID+"_"+cleanFilename)
			if _, err2 := os.Stat(tempFilePath); err2 == nil {
				localFilePath = tempFilePath
			}
		}
		if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		file, err := os.Open(localFilePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer func() { _ = file.Close() }()

		stat, err := file.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Accept-Ranges", "bytes")
		http.ServeContent(w, r, cleanFilename, stat.ModTime(), file)
	}
}
