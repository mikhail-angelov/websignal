package server

import (
	"encoding/json"
	"log"

	"github.com/gobwas/ws/wsutil"
)

const (
	createRoomMessage int = iota
	joinRoomMessage
	sdpMessage
	candidateMessage
	textMessage
)

// Message (ws) fields
type Message struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type int    `json:"type"`
	Data string `json:"data"`
}

// ProcessMessage ws messages processor
func ProcessMessage(connections map[string]*WS, from *WS, bts []byte) {
	message := Message{}
	log.Printf("Message: %s", string(bts))
	json.Unmarshal(bts, &message)
	log.Printf("Message: %s %d %s", message.Data, message.Type, message.From)
	switch message.Type {
	case textMessage:
		broadcast(connections, from.ID, message.Data)
	}
}

func broadcast(connections map[string]*WS, sender string, data string) {
	for id, ws := range connections {
		text := "> " + data
		if id == sender {
			text = "< " + data
		}
		message := Message{From: sender, Type: textMessage, Data: text, To: id}
		bts, _ := json.Marshal(message)
		var err = wsutil.WriteServerBinary(ws.Conn, bts)
		log.Printf("write message : %v, %v", message, string(bts))
		if err != nil {
			log.Printf("write message error: %s, %v", id, err)
			return
		}
	}
}
