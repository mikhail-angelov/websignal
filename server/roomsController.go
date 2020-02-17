package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/mikhail-angelov/websignal/auth"
)

type contextKey string

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
}

func (c *RoomsController) getRooms(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserInfo(r)
	if err != nil {
		log.Printf("[WARN] invalid user for %s %v", user, err)
	}
	rooms, err := c.rooms.GetUserRooms(user.ID)
	if err != nil {
		log.Printf("[WARN] cannot get rooms for %s", user.ID)
		rooms = []Room{}
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, rooms)
}
