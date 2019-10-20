package server

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"
)

// Server is http server
type Server struct {
	Port string
}

func (s *Server) composeRouter(jwtSectret string) *chi.Mux {
	var (
		rooms           = &RoomService{}
		ws              = NewWsServer(rooms)
		roomsController = NewRoomsController(rooms)
		router          = chi.NewRouter()
		auth            = NewAuth(jwtSectret)
	)
	AddFileServer(router, "/", http.Dir("./static"))
	router.HandleFunc("/ws", ws.SocketHandler)
	router.Route("/auth", auth.HTTPHandler)
	router.Route("/api", func(rapi chi.Router) {
		rapi.Group(func(r chi.Router) {
			r.Use(auth.Verifier())
			r.Use(auth.Authenticator)
			r.Route("/room", roomsController.HTTPHandler)
		})
	})

	return router
}

// Run the HTTP server
func (s *Server) Run(jwtSectret string) error {
	var (
		serve  = make(chan error, 1)
		sig    = make(chan os.Signal, 1)
		router = s.composeRouter(jwtSectret)
	)

	port := s.Port
	log.Printf("listen %s port", port)
	err := http.ListenAndServe(":"+port, router)
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
	log.Printf("[INFO] signaling server is terminated with error %+v", err)
	return err
}
