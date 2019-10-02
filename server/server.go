package server

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/rakyll/statik/fs"
)

// Server is http server
type Server struct {
	Port string
}

// Execute runs the HTTP server
func (s *Server) Execute() error {
	log.Printf("[INFO] start server on port %s", s.Port)

	err := s.run()
	if err != nil {
		log.Printf("[ERROR] remark terminated with error %+v", err)
		return err
	}
	log.Printf("[INFO] remark terminated")
	return nil
}

func (s *Server) run() error {
	var (
		serve           = make(chan error, 1)
		sig             = make(chan os.Signal, 1)
		rooms           = &RoomService{}
		ws              = NewWsServer(rooms)
		roomsController = NewRoomsController(rooms)
		r               = chi.NewRouter()
	)
	addFileServer(r, "/", http.Dir("./static"))
	r.Get("/ws", ws.SocketHandler)
	r.Get("/room", roomsController.HTTPHandler)

	port := s.Port
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("listen %s error: %v", port, err)
		return err
	}
	signal.Notify(sig, syscall.SIGTERM)

	select {
	case err := <-serve:
		log.Fatal(err)
	case sig := <-sig:
		log.Printf("signal %q received", sig)
	}
	return nil
}

// serves static files from /web or embedded by statik
func addFileServer(r chi.Router, path string, root http.FileSystem) {

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
