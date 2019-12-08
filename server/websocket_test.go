package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/mikhail-angelov/websignal/auth"
	"github.com/mikhail-angelov/websignal/logger"
)

func TestSocketHandler(t *testing.T) {
	secret := "test"
	rooms := NewRoomService()
	logger := logger.New()
	auth1 := auth.NewAuth(secret, logger, "test-url")
	wsServer := NewWsServer(rooms, auth1, logger)
	router := chi.NewRouter()
	router.HandleFunc("/ws", wsServer.SocketHandler)
	s := httptest.NewServer(router)
	defer s.Close()

	jwtService := auth.NewJWT(secret)
	claims := auth.Claims{User: &auth.User{ID: "test"}, StandardClaims: jwt.StandardClaims{
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}}
	token := jwtService.NewJwtToken(claims)

	url := "ws" + strings.TrimPrefix(s.URL, "http") + "/ws?token=" + token
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
