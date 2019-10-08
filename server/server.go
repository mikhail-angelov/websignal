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

// Run the HTTP server
func (s *Server) Run() error {
	var (
		serve           = make(chan error, 1)
		sig             = make(chan os.Signal, 1)
		rooms           = &RoomService{}
		ws              = NewWsServer(rooms)
		roomsController = NewRoomsController(rooms)
		router          = chi.NewRouter()
	)
	AddFileServer(router, "/", http.Dir("./static"))
	router.HandleFunc("/ws", ws.SocketHandler)
	router.Route("/room", roomsController.HTTPHandler)

	port := s.Port
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
