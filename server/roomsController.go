package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

//RoomResponse generic response structure
type RoomResponse struct {
	ID string `json:id`
}

//RoomRequest generic request structure
type RoomRequest struct {
	User string `json:user`
}

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
func (c *RoomsController) HTTPHandler(r chi.Router) {
	r.Post("/{owner}", c.createRoom)
	r.Delete("/{id}", c.removeRoom)
	r.Post("/join/{id}", c.joinRoom)
	r.Post("/leave/{id}", c.leaveRoom)
}

//create room
func (c *RoomsController) createRoom(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	request := &RoomRequest{}
	err = json.Unmarshal(body, request)
	log.Printf("[INFO] createRoom %s", request.User)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	room, err := c.rooms.CreateRoom(request.User)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "cannot create room")
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, &RoomResponse{ID: room.ID})
}

//remove room
func (c *RoomsController) removeRoom(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	log.Printf("[INFO] removeRoom %s", id)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error %v", err)
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	request := &RoomRequest{}
	err = json.Unmarshal(body, request)
	if err != nil {
		log.Printf("error %v", err)
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	err = c.rooms.RemoveRoom(id, request.User)
	if err != nil {
		log.Printf("error %v", err)
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "cannot remove room")
		return
	}
	render.Status(r, http.StatusOK)
	render.PlainText(w, r, "ok")
}

//join room
func (c *RoomsController) joinRoom(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	log.Printf("[INFO] joinRoom %s", id)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	request := &RoomRequest{}
	err = json.Unmarshal(body, request)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	err = c.rooms.JoinToRoom(id, request.User)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "cannot join room")
		return
	}
	render.Status(r, http.StatusOK)
	render.PlainText(w, r, "ok")
}

//leave room
func (c *RoomsController) leaveRoom(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	log.Printf("[INFO] leaveRoom %s", id)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	request := &RoomRequest{}
	err = json.Unmarshal(body, request)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	err = c.rooms.LeaveRoom(id, request.User)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "cannot leave room")
		return
	}
	render.Status(r, http.StatusOK)
	render.PlainText(w, r, "ok")
}
