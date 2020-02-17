package server

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// RoomMessage .
type RoomMessage struct {
	// set by service
	Author    string `json:"author"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

//User in room
type User struct {
	Name       string `json:"name"`
	ID         string `json:"id"`
	PeerID     string `json:"peerId"`
	Picture    []byte `json:"picture,omitempty"`
	PictureURL string `json:"pictureUrl,omitempty"`
}

//Room node
type Room struct {
	ID        string        `json:"id"`
	Owner     string        `json:"owner"`
	Users     []User        `json:"users"`
	Messages  []RoomMessage `json:"messages"`
	timestamp time.Time
}

// RoomService room service
type RoomService struct {
	rooms map[string]*Room
}

//NewRoomService create new service
func NewRoomService() *RoomService {
	return &RoomService{
		rooms: make(map[string]*Room),
	}
}

//GetRoom .
func (r *RoomService) GetRoom(id string) *Room {
	return r.rooms[id]
}

//CreateRoom creates room
func (r *RoomService) CreateRoom(owner User) (*Room, error) {
	id := uuid.New().String()
	if r.rooms[id] != nil {
		log.Printf("Room %s is already exist", id)
		return nil, errors.Errorf("already exist")
	}
	room := &Room{
		ID:        id,
		Owner:     owner.PeerID,
		Users:     []User{owner},
		Messages:  []RoomMessage{},
		timestamp: time.Now(),
	}
	r.rooms[id] = room
	return room, nil
}

//RemoveRoom remove room
func (r *RoomService) RemoveRoom(id string, owner string) error {
	if r.rooms[id] == nil {
		log.Printf("Room %s does not exist", id)
		return errors.Errorf("does not exist")
	}
	if r.rooms[id].Owner != owner {
		log.Printf("%s is not owner of room %s", owner, id)
		return errors.Errorf(owner + " is not owner")
	}
	delete(r.rooms, id)
	return nil
}

//JoinToRoom join to room
func (r *RoomService) JoinToRoom(id string, user User) (*Room, error) {
	room := r.rooms[id]
	if room == nil {
		log.Printf("Room %s does not exist", id)
		return nil, errors.Errorf("does not exist")
	}
	room.Users = append(room.Users, user)
	return room, nil
}

// LeaveRoom leave room
func (r *RoomService) LeaveRoom(roomID string, userID string) (*Room, error) {
	room := r.rooms[roomID]
	if room == nil {
		log.Printf("Room %s does not exist", roomID)
		return nil, errors.Errorf("does not exist")
	}
	room.Users = filterUsers(room.Users, func(u User) bool { return u.PeerID != userID })
	if len(r.rooms[roomID].Users) == 0 {
		r.RemoveRoom(roomID, room.Owner)
		return nil, nil
	}
	return room, nil
}

//AddFakeUser .
func (r *RoomService) AddFakeUser(roomID string, user *User) (*Room, error) {
	room := r.rooms[roomID]
	if room == nil {
		log.Printf("Room %s does not exist", roomID)
		return nil, errors.Errorf("does not exist")
	}
	room.Users = append(room.Users, *user)
	return room, nil
}

//RemoveFakeUser .
func (r *RoomService) RemoveFakeUser(roomID string, id string) (*Room, error) {
	room := r.rooms[roomID]
	if room == nil {
		log.Printf("Room %s does not exist", roomID)
		return nil, errors.Errorf("does not exist")
	}
	room.Users = filterUsers(room.Users, func(u User) bool { return u.ID != id })
	return room, nil
}

//GetUserRooms return list of rooms for particular user
func (r *RoomService) GetUserRooms(id string) ([]Room, error) {
	filtered := []Room{}
	for _, room := range r.rooms {
		if hasUser(room.Users, id) {
			filtered = append(filtered, *room)
		}
	}
	return filtered, nil
}

// RoomToMap .
func RoomToMap(room *Room) json.RawMessage {
	data := map[string]interface{}{"id": room.ID, "owner": room.Owner, "users": room.Users, "messages": room.Messages}
	bts, _ := json.Marshal(data)
	return bts
}

func filterUsers(users []User, fn func(u User) bool) []User {
	filtered := []User{}
	for _, u := range users {
		if fn(u) {
			filtered = append(filtered, u)
		}
	}
	return filtered
}

func hasUser(users []User, id string) bool {
	for _, u := range users {
		if u.PeerID == id {
			return true
		}
	}
	return false
}
