package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/mikhail-angelov/websignal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.User{ID: "test", Name: "test"}
		r = auth.SetUserInfo(r, user) // populate user info to request context
		next.ServeHTTP(w, r)
	})
}
func startupT(t *testing.T) (ts *httptest.Server, rs *RoomService, teardown func()) {

	rooms := NewRoomService()
	controller := NewRoomsController(rooms)
	router := chi.NewRouter()
	router.Route("/api", func(rapi chi.Router) {
		rapi.Group(func(r chi.Router) {
			r.Use(fakeAuth)
			r.Route("/room", controller.HTTPHandler)
		})
	})
	ts = httptest.NewServer(router)

	teardown = func() {
		ts.Close()
	}

	return ts, rooms, teardown
}
func TestGetRoomsAPI(t *testing.T) {
	ts, _, teardown := startupT(t)
	defer teardown()

	client := http.Client{}
	req, err := http.NewRequest("GET", ts.URL+"/api/room", nil)
	assert.Nil(t, err)
	resp, err := client.Do(req)

	require.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	var response []Room
	err = json.Unmarshal([]byte(body), &response)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(response))
	log.Printf("[INFO] rooms: %v ", response)
}
