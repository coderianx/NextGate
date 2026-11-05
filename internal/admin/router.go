package admin

import (
	"github.com/coderianx/NextGate/internal/admin/handlers/auth"
	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/api/v1", func(r chi.Router) {
		r.Post(
			"/login", auth.AdminLoginHandler,
		)
	})

	return router
}
