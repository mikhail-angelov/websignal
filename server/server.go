package server

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
		server = new(http.Server)
		serve  = make(chan error, 1)
		sig    = make(chan os.Signal, 1)
		ws     = NewWsServer()
	)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ws", ws.SocketHandler)

	port := s.Port
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen %s error: %v", port, err)
		return err
	}
	log.Printf("listening %s (%s)", ln.Addr(), port)

	signal.Notify(sig, syscall.SIGTERM)
	go func() { serve <- server.Serve(ln) }()

	select {
	case err := <-serve:
		log.Fatal(err)
	case sig := <-sig:
		log.Printf("signal %q received", sig)
	}
	return nil
}
