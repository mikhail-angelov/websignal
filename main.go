package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Client struct {
	Conn net.Conn
	Id   string
}

var addr = flag.String("port", ":9001", "addr to listen")

func main() {
	log.SetFlags(0)
	flag.Parse()

	clients := make(map[string]*Client)

	http.HandleFunc("/ws", helpersHighLevelHandler(clients))
	http.Handle("/", http.FileServer(http.Dir("./static")))

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %q error: %v", *addr, err)
	}
	log.Printf("listening %s (%q)", ln.Addr(), *addr)

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

func broadcast(clients map[string]*Client, sender string, bts []byte) {
	outgoingMessage := fmt.Sprintf("> %s", string(bts))
	incomingMessage := fmt.Sprintf("< %s", string(bts))
	for id, client := range clients {
		message := []byte(outgoingMessage)
		if id == sender {
			message = []byte(incomingMessage)
		}
		var err = wsutil.WriteServerText(client.Conn, message)
		if err != nil {
			log.Printf("write message error: %v", id, err)
			return
		}
	}
}

func helpersHighLevelHandler(clients map[string]*Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Printf("upgrade error: %s", err)
			return
		}
		id := r.URL.Query().Get("id")
		// todo: add auth validation here
		if id == "" {
			log.Printf("connection error, invalid id: %s", id)
			return
		}
		defer func() {
			conn.Close()
			delete(clients, id)
			broadcast(clients, id, []byte(fmt.Sprintf("%s is disconnected", id)))
		}()
		// todo: check id is used
		clients[id] = &Client{Conn: conn, Id: id}
		log.Printf("connected: %s", id)
		broadcast(clients, id, []byte(fmt.Sprintf("%s is connected", id)))

		for {
			bts, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				log.Printf("read message error: %v", err)
				return
			}
			log.Printf("read message %s , %s", bts, op)
			broadcast(clients, id, bts)
		}
	}
}
