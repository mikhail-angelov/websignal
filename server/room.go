package server

import (
	"log"

	"github.com/pkg/errors"
)

//Room is room storage
type Room struct {
	rooms map[string][]string
}

//NewRooms create new service
func NewRooms() *Room {
	res := Room{
		rooms: map[string][]string{},
	}
	return &res
}

func (r *Room) addRoom(room string, from string) error {
	if r.rooms[room] != nil {
		log.Printf("Room %s is already exist", room)
		return errors.Errorf("already exist")
	}
	r.rooms[room] = []string{}
	r.rooms[room] = append(r.rooms[room], from)
	return nil
}
