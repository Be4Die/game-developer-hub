package http

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter создаёт и настраивает chi-маршрутизатор.
func NewRouter(
	buildHandler *BuildHandler,
	instanceHandler *InstanceHandler,
	discoveryHandler *DiscoveryHandler,
	nodeHandler *NodeHandler,
	healthHandler *HealthHandler,
	_ *slog.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	// Health.
	r.Get("/health", healthHandler.Check)

	// Builds.
	r.Post("/games/{gameId}/builds", buildHandler.Upload)
	r.Get("/games/{gameId}/builds", buildHandler.List)
	r.Get("/games/{gameId}/builds/{buildVersion}", buildHandler.Get)
	r.Delete("/games/{gameId}/builds/{buildVersion}", buildHandler.Delete)

	// Instances.
	r.Post("/games/{gameId}/instances", instanceHandler.Start)
	r.Get("/games/{gameId}/instances", instanceHandler.List)
	r.Get("/games/{gameId}/instances/{instanceId}", instanceHandler.Get)
	r.Delete("/games/{gameId}/instances/{instanceId}", instanceHandler.Stop)
	r.Get("/games/{gameId}/instances/{instanceId}/logs", instanceHandler.StreamLogs)
	r.Get("/games/{gameId}/instances/{instanceId}/usage", instanceHandler.GetUsage)

	// Discovery.
	r.Get("/games/{gameId}/servers", discoveryHandler.Discover)

	// Nodes.
	r.Post("/nodes", nodeHandler.Register)
	r.Get("/nodes", nodeHandler.List)
	r.Get("/nodes/{nodeId}", nodeHandler.Get)
	r.Delete("/nodes/{nodeId}", nodeHandler.Delete)
	r.Get("/nodes/{nodeId}/usage", nodeHandler.GetUsage)

	return r
}
