package admin

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"

	adminui "github.com/coderianx/NextGate/admin-ui"
)

func newAdminUIHandler() http.Handler {
	dist := adminui.Dist()
	fileServer := http.FileServer(http.FS(dist))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setAdminUIHeaders(w, r)

		if shouldServeSPAIndex(dist, r) {
			request := r.Clone(r.Context())
			request.URL.Path = "/"
			request.URL.RawPath = ""
			fileServer.ServeHTTP(w, request)
			return
		}

		fileServer.ServeHTTP(w, r)
	})
}

func shouldServeSPAIndex(dist fs.FS, r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}

	requestPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")

	if requestPath == "" || requestPath == "." ||
		strings.HasPrefix(requestPath, "assets/") ||
		path.Ext(requestPath) != "" {
		return false
	}

	_, err := fs.Stat(dist, requestPath)

	return errors.Is(err, fs.ErrNotExist)
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
