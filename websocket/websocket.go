package websocket

import (
	"fmt"
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
			log.Printf("write message error: %s, %v", id, err)
			return
		}
	}
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
			log.Printf("read message %s , %v", bts, op)
			broadcast(clients, id, bts)
		}
	}
}
