package admin

import (
	"net/http"
	"strings"

	adminui "github.com/coderianx/NextGate/admin-ui"
)

func newAdminUIHandler() http.Handler {
	fileServer := http.FileServer(http.FS(adminui.Dist()))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setAdminUIHeaders(w, r)
		fileServer.ServeHTTP(w, r)
	})
}

func setAdminUIHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Permissions-Policy", "camera=(), geolocation=(), microphone=()")

	if strings.HasPrefix(r.URL.Path, "/assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
}
