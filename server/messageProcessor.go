package server

import (
	"encoding/json"
	"log"

	"github.com/gobwas/ws/wsutil"
)

const (
	CREATE_ROOM int = iota
	JOIN_ROOM
	SDP
	CANDIDATE
	TEXT
)

type Message struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type int    `json:"type"`
	Data string `json:"data"`
}

func ProcessMessage(clients map[string]*Client, from *Client, bts []byte) {
	message := Message{}
	log.Printf("Message: %s", string(bts))
	json.Unmarshal(bts, &message)
	log.Printf("Message: %s %d %s", message.Data, message.Type, message.From)
	switch message.Type {
	case TEXT:
		broadcast(clients, from.Id, message.Data)
	}
}

func broadcast(clients map[string]*Client, sender string, data string) {
	for id, client := range clients {
		text := "> " + data
		if id == sender {
			text = "< " + data
		}
		message := Message{From: sender, Type: TEXT, Data: text, To: id}
		bts, _ := json.Marshal(message)
		var err = wsutil.WriteServerBinary(client.Conn, bts)
		log.Printf("write message : %v, %v", message, string(bts))
		if err != nil {
			log.Printf("write message error: %s, %v", id, err)
			return
		}
	}
}
