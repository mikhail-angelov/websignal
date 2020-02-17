package server

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/mikhail-angelov/websignal/auth"
	"github.com/mikhail-angelov/websignal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	url := "ws" + strings.TrimPrefix(s.URL, "http") + "/ws?token=" + token + "&id=test"
	t.Log("ws connect: " + url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.Nil(t, err)
	defer ws.Close()
	data := composeData(map[string]interface{}{"text": "test", "id": "test"})

	message := Message{From: "sender", Type: createRoomMessage, Data: data, To: "id"}
	bts, _ := json.Marshal(message)
	err = ws.WriteMessage(websocket.TextMessage, bts)
	assert.Nil(t, err)
	_, p, err := ws.ReadMessage()
	assert.Nil(t, err)
	msg := Message{}
	json.Unmarshal(p, &msg)
	messageData := InputMessageData{}
	json.Unmarshal(msg.Data, &messageData)

	text := fmt.Sprintf("%v", messageData["owner"])
	require.Equal(t, "test", text)
}
