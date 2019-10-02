package server

import (
	"log"

	"github.com/pkg/errors"
)

//Room node
type Room struct {
	id    string
	owner string
	users []string
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
func (r *RoomService) CreateRoom(id string, owner string) error {
	if r.rooms[id] != nil {
		log.Printf("Room %s is already exist", id)
		return errors.Errorf("already exist")
	}
	r.rooms[id] = &Room{
		id:    id,
		owner: owner,
		users: []string{owner},
	}
	return nil
}

//RemoveRoom remove room
func (r *RoomService) RemoveRoom(id string, owner string) error {
	if r.rooms[id] == nil {
		log.Printf("Room %s does not exist", id)
		return errors.Errorf("does not exist")
	}
	if r.rooms[id].owner != owner {
		log.Printf("%s is not owner of room %s", owner, id)
		return errors.Errorf("is not owner")
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
	r.rooms[id].users = append(r.rooms[id].users, user)
	return nil
}

//LeaveRoom leave room
func (r *RoomService) LeaveRoom(id string, user string) error {
	if r.rooms[id] == nil {
		log.Printf("Room %s does not exist", id)
		return errors.Errorf("does not exist")
	}
	r.rooms[id].users = filterUsers(r.rooms[id].users, func(u string) bool { return u != user })
	if len(r.rooms[id].users) == 0 {
		r.RemoveRoom(id, r.rooms[id].owner)
	}
	return nil
}

//GetRoomUsers room users
func (r *RoomService) GetRoomUsers(id string) ([]string, error) {
	if r.rooms[id] == nil {
		log.Printf("Room %s does not exist", id)
		return nil, errors.Errorf("does not exist")
	}
	return r.rooms[id].users, nil
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
