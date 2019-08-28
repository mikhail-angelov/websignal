package websocket

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"encoding/json"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Client struct {
	Conn net.Conn
	Id   string
}

type Message struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
	Data string `json:"data"`
}

func broadcast(clients map[string]*Client, sender string, data string) {
	for id, client := range clients {
		text := "> " + data
		if id == sender {
			text = "< " + data
		}
		message := Message{From: sender, Type: "text", Data: text, To: id}
		bts, _ := json.Marshal(message)
		var err = wsutil.WriteServerBinary(client.Conn, bts)
		log.Printf("write message : %s, %v", message, string(bts))
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
			broadcast(clients, id, fmt.Sprintf("%s is disconnected", id))
		}()
		// todo: check id is used
		clients[id] = &Client{Conn: conn, Id: id}
		log.Printf("connected: %s", id)
		broadcast(clients, id, fmt.Sprintf("%s is connected", id))

		for {
			bts, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				log.Printf("read message error: %v", err)
				return
			}
			message := Message{}
			json.Unmarshal(bts, &message)
			log.Printf("Message: %s %s %s", message.Data, message.Type, message.From)
			broadcast(clients, id, message.Data)
		}
	}
}
