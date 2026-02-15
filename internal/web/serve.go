package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var staticFS embed.FS

// FS returns the filesystem for the web UI (rooted at static/).
func FS() (fs.FS, error) {
	return fs.Sub(staticFS, "static")
}

// MustFS returns the embedded static FS rooted at static. Panics on error.
func MustFS() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return sub
}

// SPAHandler serves the embedded files and falls back to index.html for non-file paths (SPA routing).
func SPAHandler(apiHandler http.Handler) http.Handler {
	fsys := MustFS()
	fileServer := http.FileServer(http.FS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
			apiHandler.ServeHTTP(w, r)
			return
		}
		// Root: serve index.html
		if r.URL.Path == "" || r.URL.Path == "/" {
			idx, _ := fs.ReadFile(fsys, "index.html")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(idx)
			return
		}
		// Try to serve a static file (e.g. /manifest.webmanifest)
		name := r.URL.Path[1:]
		if name != "" && name[0] != '.' {
			f, err := fsys.Open(name)
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		// SPA fallback: /tasks, /report, etc.
		idx, _ := fs.ReadFile(fsys, "index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(idx)
	})
}
