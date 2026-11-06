// Package api configures the public HTTP API.
package api

import (
	"github.com/coderianx/NextGate/internal/api/handlers/apikeys"
	"github.com/coderianx/NextGate/internal/api/handlers/auth"
	"github.com/coderianx/NextGate/internal/middlewares"
	"github.com/go-chi/chi/v5"
)

// NewRouter returns the public API router with authentication endpoints mounted
// under /api/v1/auth.
func NewRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", auth.RegisterHandler)
			r.Post("/login", auth.LoginHandler)
			r.Post("/refresh", auth.RefreshHandler)
		})

		r.With(
			middlewares.AuthMiddleware("user"),
		).Post(
			"/apikeys", apikeys.CreateNewAPIKey,
		)

		r.With(
			middlewares.AuthMiddleware("user"),
		).Get("/apikeys", apikeys.ListMyAPIKeys)

		r.With(
			middlewares.AuthMiddleware("user"),
		).Delete("/apikeys/{id}", apikeys.DeleteAPIKey)
	})

	return router
}
