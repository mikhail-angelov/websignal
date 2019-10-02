package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestSocketHandler(t *testing.T) {
	rooms := NewRoomService()
	wsServer := NewWsServer(rooms)
	s := httptest.NewServer(http.HandlerFunc(wsServer.SocketHandler))
	defer s.Close()

	url := "ws" + strings.TrimPrefix(s.URL, "http") + "/?id=test"
	t.Log("ws connect: " + url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	message := Message{From: "sender", Type: textMessage, Data: "test", To: "id"}
	bts, _ := json.Marshal(message)
	if err := ws.WriteMessage(websocket.TextMessage, bts); err != nil {
		t.Fatalf("%v", err)
	}
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	json.Unmarshal(p, &message)
	if string(message.Data) != "< test" {
		t.Fatalf("bad message")
	}
}
