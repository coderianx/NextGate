// Package admin configures the authenticated HTTP API used to manage the
// NextGate installation.
package admin

import (
	"github.com/coderianx/NextGate/internal/admin/handlers/auth"
	"github.com/coderianx/NextGate/internal/admin/handlers/providers"
	"github.com/coderianx/NextGate/internal/middlewares"
	"github.com/go-chi/chi/v5"
)

// NewRouter returns the admin HTTP router with login, provider, and model
// management endpoints mounted under /api/v1.
func NewRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/api/v1", func(r chi.Router) {
		r.Post(
			"/login", auth.AdminLoginHandler,
		)

		r.With(
			middlewares.AuthMiddleware("admin"),
		).Post("/providers", providers.AddNewProvider)

		r.With(
			middlewares.AuthMiddleware("admin"),
		).Get(
			"/providers", providers.ListMyProviders,
		)

		r.With(
			middlewares.AuthMiddleware("admin"),
		).Post(
			"/providers/{id}/models", providers.AddNewModel,
		)

		r.With(
			middlewares.AuthMiddleware("admin"),
		).Get(
			"/providers/{id}/models", providers.ListModels,
		)

		r.With(
			middlewares.AuthMiddleware("admin"),
		).Patch(
			"/providers/{id}/models/{model_id}", providers.SetModelEnabled,
		)

		r.With(
			middlewares.AuthMiddleware("admin"),
		).Delete(
			"/providers/{id}/models/{model_id}", providers.DeleteModel,
		)

		r.With(
			middlewares.AuthMiddleware("admin"),
		).Delete(
			"/providers/{id}", providers.DeleteProvider,
		)
	})

	router.Handle("/*", newAdminUIHandler())

	return router
}
