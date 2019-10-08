package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startupT(t *testing.T) (ts *httptest.Server, rs *RoomService, teardown func()) {

	rooms := NewRoomService()
	controller := NewRoomsController(rooms)
	router := chi.NewRouter()
	router.Route("/room", controller.HTTPHandler)
	ts = httptest.NewServer(router)

	teardown = func() {
		ts.Close()
	}

	return ts, rooms, teardown
}
func TestCreateRoomAPI(t *testing.T) {
	ts, _, teardown := startupT(t)
	defer teardown()

	r := strings.NewReader(`{"id":"test"}`)
	client := http.Client{}
	req, err := http.NewRequest("POST", ts.URL+"/room/test", r)
	assert.Nil(t, err)
	resp, err := client.Do(req)

	require.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	response := &RoomResponse{}
	err = json.Unmarshal(body, response)
	assert.Nil(t, err)
	log.Printf("[INFO] created id: %s ", response.ID)
}

func TestRemoveRoomRoomAPI(t *testing.T) {
	ts, rooms, teardown := startupT(t)
	defer teardown()

	user := "test"
	room, err := rooms.CreateRoom(user)
	assert.Nil(t, err)

	r := strings.NewReader(fmt.Sprintf(`{"user":"%s"}`, user))
	client := http.Client{}
	req, err := http.NewRequest("DELETE", ts.URL+"/room/"+room.ID, r)
	assert.Nil(t, err)
	resp, err := client.Do(req)

	require.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	require.Equal(t, string(body), "ok")
}

func TestJoinRoomRoomAPI(t *testing.T) {
	ts, rooms, teardown := startupT(t)
	defer teardown()

	user := "test"
	other := "test2"
	room, err := rooms.CreateRoom(user)
	assert.Nil(t, err)

	r := strings.NewReader(fmt.Sprintf(`{"user":"%s"}`, other))
	client := http.Client{}
	req, err := http.NewRequest("POST", ts.URL+"/room/join/"+room.ID, r)
	assert.Nil(t, err)
	resp, err := client.Do(req)

	require.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	require.Equal(t, string(body), "ok")
	users, err := rooms.GetRoomUsers(room.ID)
	assert.Nil(t, err)
	assert.Equal(t, len(users), 2)
}
