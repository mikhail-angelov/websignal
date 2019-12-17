package server

import (
	"encoding/json"
	"fmt"
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
	socketID := r.URL.Query().Get("id")
	id, err := s.auth.ValidateToken(token)
	if err != nil || id == "" || socketID == "" {
		s.log.Logf("[WARN] connection error, invalid id:  %s, %s, %v", id, socketID, err)
		return
	}
	defer delete(s.clients, socketID)
	// todo: check id is used
	client := &WS{Conn: conn, ID: id}
	s.clients[socketID] = client
	s.log.Logf("[INFO] connected: %s", id)

	for {
		bts, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			s.log.Logf("[WARN] read message error:  %v", err)
			return
		}
		s.processMessage(client, socketID, bts)
	}
}

func (s *WsServer) processMessage(from *WS, socketID string, bts []byte) error {
	var err error
	message := Message{}
	//log.Printf("Message: %s", string(bts))
	json.Unmarshal(bts, &message)
	log.Printf("receive %d from %s to %s", message.Type, message.From, message.To)
	switch message.Type {
	case textMessage:
		broadcast(s.clients, from.ID, message.Data)
	case createRoomMessage:
		room, err := s.rooms.CreateRoom(socketID)
		if err != nil {
			return errors.Errorf("create room error")
		}
		to := s.clients[socketID]
		data := map[string]interface{}{"id": room.ID, "owner": room.Owner}
		err = send(to.Conn, &Message{From: socketID, Type: roomIsCreatedMessage, Data: data, To: socketID})
	case joinRoomMessage:
		roomID := fmt.Sprintf("%v", message.Data["id"])
		room, err := s.rooms.JoinToRoom(roomID, socketID)
		if err != nil {
			return errors.Errorf("create room error")
		}
		to := s.clients[socketID]
		data := map[string]interface{}{"id": room.ID, "owner": room.Owner}
		err = send(to.Conn, &Message{From: socketID, Type: roomToJoinMessage, Data: data, To: socketID})
	case leaveRoomMessage:
		roomID := fmt.Sprintf("%v", message.Data["id"])
		err = s.rooms.LeaveRoom(roomID, socketID)
	case sdpMessage:
		to := s.clients[message.To]
		if to == nil {
			return errors.Errorf("invalid destination")
		}
		err = send(to.Conn, &Message{From: socketID, Type: sdpMessage, Data: message.Data, To: message.To})
	case candidateMessage:
		to := s.clients[message.To]
		if to == nil {
			return errors.Errorf("invalid destination")
		}
		err = send(to.Conn, &Message{From: socketID, Type: candidateMessage, Data: message.Data, To: message.To})
	}
	return err
}

func broadcast(connections map[string]*WS, sender string, data map[string]interface{}) {
	log.Printf("broadcast : %d %v", len(connections), data)
	for id, ws := range connections {
		message := Message{From: sender, Type: textMessage, Data: data, To: id}
		err := send(ws.Conn, &message)
		if err != nil {
			log.Printf("send message error: %s, %v", id, err)
		}
	}
}

func send(conn net.Conn, message *Message) error {
	log.Printf("send %d to %s", message.Type, message.To)
	bts, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return wsutil.WriteServerBinary(conn, bts)
}
