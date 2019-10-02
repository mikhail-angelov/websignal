package server

import (
	"log"
	"net/http"
)

//RoomsController controlles structure
type RoomsController struct {
	rooms *RoomService
}

//NewRoomsController constructor
func NewRoomsController(rooms *RoomService) *RoomsController {
	return &RoomsController{
		rooms: rooms,
	}
}

//HTTPHandler main handler
func (s *RoomsController) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] http %v", r)
	if r.Method == "POST" {
		//create room
		log.Printf("[INFO] http %s, %v", r.RequestURI, r.Body)

	} else if r.Method == "DELETE" {
		//remove room
	}
}
