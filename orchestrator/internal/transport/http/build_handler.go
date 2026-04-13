package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	"github.com/go-chi/chi/v5"
)

// BuildHandler обрабатывает HTTP-запросы для управления билдами.
type BuildHandler struct {
	pipeline *service.BuildPipeline
}

// NewBuildHandler создаёт обработчик билдов.
func NewBuildHandler(pipeline *service.BuildPipeline) *BuildHandler {
	return &BuildHandler{pipeline: pipeline}
}

// Upload обрабатывает POST /games/{gameId}/builds.
func (h *BuildHandler) Upload(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		badRequest(w, "invalid gameId")
		return
	}

	// Размер multipart-формы ограничен на уровне сервера (config.Limits.MaxBuildSizeBytes).
	//nolint:gosec
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		badRequest(w, "failed to parse multipart form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		badRequest(w, "missing or invalid 'image' file: "+err.Error())
		return
	}
	defer func() { _ = file.Close() }()

	//nolint:gosec // FormValue безопасен после ParseMultipartForm
	version := r.FormValue("build_version")
	if version == "" {
		badRequest(w, "missing 'build_version'")
		return
	}

	protocolStr := r.FormValue("protocol") //nolint:gosec
	protocol := parseProtocol(protocolStr)
	if protocol == 0 {
		badRequest(w, "invalid or missing 'protocol'")
		return
	}

	internalPortStr := r.FormValue("internal_port") //nolint:gosec
	internalPort, err := strconv.ParseUint(internalPortStr, 10, 32)
	if err != nil || internalPort == 0 {
		badRequest(w, "invalid or missing 'internal_port'")
		return
	}

	maxPlayersStr := r.FormValue("max_players") //nolint:gosec
	maxPlayers, err := strconv.ParseUint(maxPlayersStr, 10, 32)
	if err != nil || maxPlayers == 0 {
		badRequest(w, "invalid or missing 'max_players'")
		return
	}

	params := service.UploadBuildParams{
		GameID:       gameID,
		Version:      version,
		Protocol:     protocol,
		InternalPort: uint32(internalPort),
		MaxPlayers:   uint32(maxPlayers),
		Archive:      file,
		ArchiveSize:  header.Size,
	}

	build, err := h.pipeline.UploadBuild(r.Context(), params)
	if handleDomainError(w, err, "upload build") {
		return
	}

	jsonOK(w, buildResponse(build), http.StatusCreated)
}

// List обрабатывает GET /games/{gameId}/builds.
func (h *BuildHandler) List(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		badRequest(w, "invalid gameId")
		return
	}

	builds, err := h.pipeline.ListBuilds(r.Context(), gameID)
	if handleDomainError(w, err, "list builds") {
		return
	}

	resp := make([]any, 0, len(builds))
	for _, b := range builds {
		resp = append(resp, buildResponse(b))
	}

	jsonOK(w, map[string]any{"builds": resp}, http.StatusOK)
}

// Get обрабатывает GET /games/{gameId}/builds/{buildVersion}.
func (h *BuildHandler) Get(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		badRequest(w, "invalid gameId")
		return
	}

	version := chi.URLParam(r, "buildVersion")
	if version == "" {
		badRequest(w, "missing buildVersion")
		return
	}

	build, err := h.pipeline.GetBuild(r.Context(), gameID, version)
	if handleDomainError(w, err, "get build") {
		return
	}

	jsonOK(w, buildResponse(build), http.StatusOK)
}

// Delete обрабатывает DELETE /games/{gameId}/builds/{buildVersion}.
func (h *BuildHandler) Delete(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		badRequest(w, "invalid gameId")
		return
	}

	version := chi.URLParam(r, "buildVersion")
	if version == "" {
		badRequest(w, "missing buildVersion")
		return
	}

	err = h.pipeline.DeleteBuild(r.Context(), gameID, version)
	if err != nil {
		if errors.Is(err, domain.ErrBuildInUse) {
			conflict(w, err.Error())
			return
		}
		if handleDomainError(w, err, "delete build") {
			return
		}
	}

	jsonOK204(w)
}

// buildResponse преобразует доменный билд в API-ответ.
func buildResponse(b *domain.ServerBuild) map[string]any {
	return map[string]any{
		"id":              b.ID,
		"game_id":         b.GameID,
		"build_version":   b.Version,
		"image_tag":       b.ImageTag,
		"protocol":        protocolResponse(b.Protocol),
		"internal_port":   b.InternalPort,
		"max_players":     b.MaxPlayers,
		"file_size_bytes": b.FileSize,
		"created_at":      b.CreatedAt,
	}
}
