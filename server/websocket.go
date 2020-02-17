package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

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
	authUser, err := s.auth.ValidateToken(token)
	id := authUser.ID
	if err != nil || id == "" || socketID == "" {
		s.log.Logf("[WARN] connection auth error:  %s, %s, %s", id, socketID, err)
		return
	}
	// continue connection after validation
	defer delete(s.clients, socketID)
	// todo: check id is used
	client := &WS{Conn: conn, ID: id}
	s.clients[socketID] = client
	s.log.Logf("[INFO] connected: %s", id)
	user := User{ID: authUser.ID, PeerID: socketID, Name: authUser.Name, Picture: authUser.Picture, PictureURL: authUser.PictureURL}

	for {
		bts, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			s.log.Logf("[WARN] read message error:  %v", err)
			s.onCloseConnection(user)
			return
		}
		s.processMessage(client, socketID, user, bts)
	}
}

func (s *WsServer) processMessage(from *WS, socketID string, user User, bts []byte) error {
	var err error
	message := Message{}
	//log.Printf("Message: %s", string(bts))
	json.Unmarshal(bts, &message)
	messageData := InputMessageData{}
	json.Unmarshal(message.Data, &messageData)

	log.Printf("receive %d from %s to %s", message.Type, socketID, message.To)
	switch message.Type {
	case textMessage:
		roomID := messageData["id"]
		text := messageData["text"]
		room := s.rooms.GetRoom(roomID)
		log.Printf("on text message at room %s", roomID)
		if room == nil {
			return errors.Errorf("send message error, no room  %s", roomID)
		}
		newMessage := RoomMessage{Author: user.PeerID, Text: text, Timestamp: time.Now().String()}
		room.Messages = append(room.Messages, newMessage)
		data := composeData(map[string]interface{}{"timestamp": newMessage.Timestamp, "author": newMessage.Author, "text": text})
		msg := &Message{From: socketID, Type: textMessage, Data: data, To: socketID}
		for _, user := range room.Users {
			to := s.clients[user.PeerID]
			err = send(to, msg)
		}
	case createRoomMessage:
		room, err := s.rooms.CreateRoom(user)
		if err != nil {
			return errors.Errorf("create room error")
		}
		to := s.clients[socketID]
		data := RoomToMap(room)
		err = send(to, &Message{From: socketID, Type: roomIsCreatedMessage, Data: data, To: socketID})
	case joinRoomMessage:
		roomID := messageData["id"]
		peerID := messageData["peerId"]
		room, err := s.rooms.JoinToRoom(roomID, user)
		if err != nil || peerID == "" {
			return errors.Errorf("join room error %s %s %v", roomID, message.To, err)
		}
		masterPeer := room.Owner //temp, all peers connects to room owner
		to := s.clients[masterPeer]
		data := composeData(map[string]interface{}{"peerId": socketID})
		err = send(to, &Message{From: socketID, Type: startPeerConnectionMessage, Data: data, To: message.To})
		if err != nil {
			return errors.Errorf("join room error cannot send start connect message %s %s %v", masterPeer, roomID, err)
		}
		log.Printf("joinRoomMessage to %s %s", roomID, message.To)
		data = RoomToMap(room)
		msg := &Message{From: socketID, Type: roomUpdateMessage, Data: data, To: message.To}
		err = s.sendToAllRoom(room, msg)
	case leaveRoomMessage:
		roomID := messageData["id"]
		room, err := s.rooms.LeaveRoom(roomID, socketID)
		if err != nil || room == nil {
			return errors.Errorf("leave room error %s %s", roomID, err)
		}
		log.Printf("leaveRoomMessage to %s %s", roomID, message.To)
		data := RoomToMap(room)
		msg := &Message{From: socketID, Type: roomUpdateMessage, Data: data, To: message.To}
		err = s.sendToRoom(room, msg, socketID)
	case addFakeUser:
		roomID := messageData["roomId"]
		id := messageData["id"]
		log.Printf("addFakeUser to %s %s", roomID, id)
		room, err := s.rooms.AddFakeUser(roomID, &User{ID: id, Name: messageData["name"], PictureURL: messageData["pictureUrl"]})
		if room == nil || err != nil {
			return errors.Errorf("add fake user to room error %s", roomID)
		}
		data := RoomToMap(room)
		msg := &Message{From: socketID, Type: roomUpdateMessage, Data: data, To: message.To}
		err = s.sendToAllRoom(room, msg)
	case removeFakeUser:
		roomID := messageData["roomId"]
		id := messageData["id"]
		log.Printf("removeFakeUser to %s %s", roomID, id)
		room, err := s.rooms.RemoveFakeUser(roomID, id)
		if room == nil || err != nil {
			return errors.Errorf("remove fake user to room error %s", roomID)
		}
		data := RoomToMap(room)
		msg := &Message{From: socketID, Type: roomUpdateMessage, Data: data, To: message.To}
		err = s.sendToAllRoom(room, msg)
	case sdpMessage:
		to := s.clients[message.To]
		err = send(to, &Message{From: socketID, Type: sdpMessage, Data: message.Data, To: message.To})
	case candidateMessage:
		to := s.clients[message.To]
		err = send(to, &Message{From: socketID, Type: candidateMessage, Data: message.Data, To: message.To})
	}
	return err
}

func (s *WsServer) sendToAllRoom(room *Room, msg *Message) error {
	var err error = nil
	for _, user := range room.Users {
		to := s.clients[user.PeerID]
		err = send(to, msg)
	}
	return err
}
func (s *WsServer) sendToRoom(room *Room, msg *Message, origin string) error {
	var err error = nil
	for _, user := range room.Users {
		if user.PeerID != origin {
			to := s.clients[user.PeerID]
			err = send(to, msg)
		}
	}
	return err
}

func (s *WsServer) onCloseConnection(user User) {
	rooms, err := s.rooms.GetUserRooms(user.PeerID)
	if err != nil {
		log.Printf("onCloseConnection no rooms for %s", user.PeerID)
		return
	}
	log.Printf("onCloseConnection leave %d rooms for %s", len(rooms), user.PeerID)
	for _, room := range rooms {
		log.Printf("leaveRoomMessage to %s %s", room.ID, user.PeerID)
		updatedRoom, err := s.rooms.LeaveRoom(room.ID, user.PeerID)
		if err != nil {
			log.Printf("leave room error %s %s", room.ID, err)
			return
		}
		if updatedRoom == nil {
			log.Printf("room is blank and removed %s", room.ID)
			return
		}
		data := RoomToMap(updatedRoom)
		for _, u := range updatedRoom.Users {
			if u.PeerID != user.PeerID {
				to := s.clients[u.PeerID]
				err = send(to, &Message{From: "offline", Type: roomUpdateMessage, Data: data, To: "all"})
			}
		}
	}
}

func send(to *WS, message *Message) error {
	log.Printf("send %d to %s", message.Type, message.To)
	bts, err := json.Marshal(message)
	if err != nil || to == nil || to.Conn == nil {
		return err
	}
	return wsutil.WriteServerBinary(to.Conn, bts)
}

func composeData(data map[string]interface{}) json.RawMessage {
	bts, _ := json.Marshal(data)
	return bts
}
