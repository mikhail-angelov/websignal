package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/rakyll/statik/fs"
)

// AddFileServer serves static files from /web or embedded by statik
func AddFileServer(r chi.Router, path string, root http.FileSystem) {

	var webFS http.Handler

	statikFS, err := fs.New()
	if err != nil {
		log.Printf("[DEBUG] no embedded assets loaded, %s", err)
		log.Printf("[INFO] run file server for %s, path %s", root, path)
		webFS = http.FileServer(root)
	} else {
		log.Printf("[INFO] run file server for %s, embedded", root)
		webFS = http.FileServer(statikFS)
	}

	origPath := path
	webFS = http.StripPrefix(path, webFS)
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		// don't show dirs, just serve files
		if strings.HasSuffix(r.URL.Path, "/") && len(r.URL.Path) > 1 && r.URL.Path != (origPath+"/") {
			http.NotFound(w, r)
			return
		}
		webFS.ServeHTTP(w, r)
	})
}
