package server

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/the127/hivetrack/web"
)

// spaHandler serves the embedded Vue SPA. Any path that doesn't match a real
// static asset falls back to index.html so client-side routing works.
func spaHandler() http.Handler {
	distFS, err := fs.Sub(web.Assets, "dist")
	if err != nil {
		panic("web assets unavailable: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if f, err := distFS.Open(path); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Unknown path — serve index.html for client-side routing.
		index, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			http.Error(w, "app not available", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(index)
	})
}
