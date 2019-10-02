package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pkg/errors"
)

// WS is websocket connection
type WS struct {
	Conn net.Conn
	ID   string
}

// WsServer is websocket server
type WsServer struct {
	clients map[string]*WS
	rooms   *RoomService
}

//NewWsServer create new service
func NewWsServer(rooms *RoomService) *WsServer {
	res := WsServer{
		clients: make(map[string]*WS),
		rooms:   rooms,
	}
	return &res
}

// SocketHandler process ws messages
func (s *WsServer) SocketHandler(w http.ResponseWriter, r *http.Request) {
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
		delete(s.clients, id)
	}()
	// todo: check id is used
	client := &WS{Conn: conn, ID: id}
	s.clients[id] = client
	log.Printf("connected: %s", id)

	for {
		bts, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			log.Printf("read message error: %v", err)
			return
		}
		s.processMessage(client, bts)
	}
}

func (s *WsServer) processMessage(from *WS, bts []byte) error {
	var err error
	message := Message{}
	log.Printf("Message: %s", string(bts))
	json.Unmarshal(bts, &message)
	log.Printf("Message: %s %d %s", message.Data, message.Type, message.From)
	switch message.Type {
	case textMessage:
		broadcast(s.clients, from.ID, message.Data)
	case joinRoomMessage:
		err = s.rooms.JoinToRoom(message.Data, from.ID)
	case leaveRoomMessage:
		err = s.rooms.LeaveRoom(message.Data, from.ID)
	case sdpMessage:
		to := s.clients[message.To]
		if to == nil {
			return errors.Errorf("invalid destination")
		}
		err = send(to.Conn, &Message{From: from.ID, Type: sdpMessage, Data: message.Data, To: message.To})
	case candidateMessage:
		to := s.clients[message.To]
		if to == nil {
			return errors.Errorf("invalid destination")
		}
		err = send(to.Conn, &Message{From: from.ID, Type: candidateMessage, Data: message.Data, To: message.To})
	}
	return err
}

func broadcast(connections map[string]*WS, sender string, data string) {
	for id, ws := range connections {
		text := "> " + data
		if id == sender {
			text = "< " + data
		}
		message := Message{From: sender, Type: textMessage, Data: text, To: id}
		var err = send(ws.Conn, &message)
		log.Printf("write message : %v", message)
		if err != nil {
			log.Printf("write message error: %s, %v", id, err)
			return
		}
	}
}

func send(conn net.Conn, message *Message) error {
	bts, err := json.Marshal(&message)
	if err != nil {
		return err
	}
	return wsutil.WriteServerBinary(conn, bts)
}
