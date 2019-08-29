package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mikhail-angelov/websignal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9001"
	}

	http.HandleFunc("/ws", server.SocketHandler())
	http.Handle("/", http.FileServer(http.Dir("./static")))

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen %s error: %v", port, err)
	}
	log.Printf("listening %s (%s)", ln.Addr(), port)

	var (
		s     = new(http.Server)
		serve = make(chan error, 1)
		sig   = make(chan os.Signal, 1)
	)
	signal.Notify(sig, syscall.SIGTERM)
	go func() { serve <- s.Serve(ln) }()

	select {
	case err := <-serve:
		log.Fatal(err)
	case sig := <-sig:
		const timeout = 5 * time.Second

		log.Printf("signal %q received; shutting down with %s timeout", sig, timeout)

		ctx, _ := context.WithTimeout(context.Background(), timeout)
		if err := s.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}
}
