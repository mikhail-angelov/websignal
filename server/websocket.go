package server

import (
	"log"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Client struct {
	Conn net.Conn
	Id   string
}

func SocketHandler() func(w http.ResponseWriter, r *http.Request) {
	clients := make(map[string]*Client)
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
		}()
		// todo: check id is used
		client := &Client{Conn: conn, Id: id}
		clients[id] = client
		log.Printf("connected: %s", id)

		for {
			bts, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				log.Printf("read message error: %v", err)
				return
			}
			ProcessMessage(clients, client, bts)
		}
	}
}
