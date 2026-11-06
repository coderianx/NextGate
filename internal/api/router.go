// Package api configures the public HTTP API.
package api

import (
	"github.com/coderianx/NextGate/internal/api/handlers/auth"
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
	})

	return router
}
