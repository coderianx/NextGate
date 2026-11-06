package admin

import (
	"github.com/coderianx/NextGate/internal/admin/handlers/auth"
	"github.com/coderianx/NextGate/internal/admin/handlers/providers"
	"github.com/coderianx/NextGate/internal/middlewares"
	"github.com/go-chi/chi/v5"
)

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
		).Delete(
			"/providers/{id}", providers.DeleteProvider,
		)
	})

	return router
}
