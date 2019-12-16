package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/mikhail-angelov/websignal/auth"
)

type contextKey string

//RoomResponse generic response structure
type RoomResponse struct {
	ID string `json:"id"`
}

//RoomRequest generic request structure
type RoomRequest struct {
	User string `json:"user"`
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
	r.Get("/", c.getRooms)
	r.Post("/", c.createRoom)
	r.Delete("/{id}", c.removeRoom)
	r.Post("/join/{id}", c.joinRoom)
	r.Post("/leave/{id}", c.leaveRoom)
}

func (c *RoomsController) getRooms(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserInfo(r)
	if err != nil {
		log.Printf("[WARN] invalid user for %s %v", user, err)
	}
	rooms, err := c.rooms.GetRoomUsers(user.ID)
	if err != nil {
		log.Printf("[WARN] cannot get rooms for %s", user.ID)
		rooms = []string{}
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, rooms)
}

//create room
func (c *RoomsController) createRoom(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserInfo(r)
	if err != nil {
		log.Printf("[WARN] invalid user for %s %v", user, err)
	}
	log.Printf("[INFO] createRoom for %s", user.ID)
	room, err := c.rooms.CreateRoom(user.ID)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrorResponse{Error: "cannot create room"})
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, &room)
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
	user, err := auth.GetUserInfo(r)
	id := chi.URLParam(r, "id")
	log.Printf("[INFO] joinRoom %s for user %s", id, user.ID)
	room, err := c.rooms.JoinToRoom(id, user.ID)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrorResponse{Error: "cannot join room"})
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, &room)
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
