package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/mikhail-angelov/websignal/auth"
	"github.com/mikhail-angelov/websignal/logger"
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
	auth    *auth.Auth
	log     *logger.Log
}

//NewWsServer create new service
func NewWsServer(rooms *RoomService, auth *auth.Auth, log *logger.Log) *WsServer {
	res := WsServer{
		clients: make(map[string]*WS),
		rooms:   rooms,
		auth:    auth,
		log:     log,
	}
	return &res
}

// SocketHandler process ws messages
func (s *WsServer) SocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		s.log.Logf("[WARN] upgrade error: %s", err)
		return
	}
	defer conn.Close()
	token := r.URL.Query().Get("token")
	id, err := s.auth.ValidateToken(token)
	if err != nil || id == "" {
		s.log.Logf("[WARN] connection error, invalid id:  %s, %v", id, err)
		return
	}
	defer delete(s.clients, id)
	// todo: check id is used
	client := &WS{Conn: conn, ID: id}
	s.clients[id] = client
	s.log.Logf("[INFO] connected: %s", id)

	for {
		bts, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			s.log.Logf("[WARN] read message error:  %v", err)
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
	log.Printf("broadcast : %d %v", len(connections), data)
	for id, ws := range connections {
		text := "> " + data
		if id == sender {
			text = "< " + data
		}
		message := Message{From: sender, Type: textMessage, Data: text, To: id}
		err := send(ws.Conn, &message)
		if err != nil {
			log.Printf("send message error: %s, %v", id, err)
		}
	}
}

func send(conn net.Conn, message *Message) error {
	log.Printf("send : %v", message)
	bts, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return wsutil.WriteServerBinary(conn, bts)
}
