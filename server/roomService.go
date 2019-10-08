package server

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

//Room node
type Room struct {
	ID        string
	Owner     string
	Users     []string
	Timestamp time.Time
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

//CreateRoom creates room
func (r *RoomService) CreateRoom(owner string) (*Room, error) {
	id := uuid.New().String()
	if r.rooms[id] != nil {
		log.Printf("Room %s is already exist", id)
		return nil, errors.Errorf("already exist")
	}
	r.rooms[id] = &Room{
		ID:        id,
		Owner:     owner,
		Users:     []string{owner},
		Timestamp: time.Now(),
	}
	return r.rooms[id], nil
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
func (r *RoomService) JoinToRoom(id string, user string) error {
	if r.rooms[id] == nil {
		log.Printf("Room %s does not exist", id)
		return errors.Errorf("does not exist")
	}
	r.rooms[id].Users = append(r.rooms[id].Users, user)
	return nil
}

//LeaveRoom leave room
func (r *RoomService) LeaveRoom(id string, user string) error {
	if r.rooms[id] == nil {
		log.Printf("Room %s does not exist", id)
		return errors.Errorf("does not exist")
	}
	r.rooms[id].Users = filterUsers(r.rooms[id].Users, func(u string) bool { return u != user })
	if len(r.rooms[id].Users) == 0 {
		r.RemoveRoom(id, r.rooms[id].Owner)
	}
	return nil
}

//GetRoomUsers room users
func (r *RoomService) GetRoomUsers(id string) ([]string, error) {
	if r.rooms[id] == nil {
		log.Printf("Room %s does not exist", id)
		return nil, errors.Errorf("does not exist")
	}
	return r.rooms[id].Users, nil
}

func filterUsers(users []string, fn func(u string) bool) []string {
	filtered := []string{}
	for _, u := range users {
		if fn(u) {
			filtered = append(filtered, u)
		}
	}
	return filtered
}
