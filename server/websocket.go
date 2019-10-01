package server

import (
	"log"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// WS is websocket connection
type WS struct {
	Conn net.Conn
	ID   string
}

// WsServer is websocket server
type WsServer struct {
	clients map[string]*WS
}

//NewWsServer create new service
func NewWsServer() *WsServer {
	res := WsServer{
		clients: map[string]*WS{},
	}
	return &res
}

// SocketHandler process ws messages
func (s *WsServer) SocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Printf("upgrade error: %s", err)
		return
	}
	id := r.URL.Query().Get("id")
	// todo: add auth validation here
	if id == "" {
		log.Printf("connection error, invalid id: %s", id)
		return
	}
	defer func() {
		conn.Close()
		delete(s.clients, id)
	}()
	// todo: check id is used
	client := &WS{Conn: conn, ID: id}
	s.clients[id] = client
	log.Printf("connected: %s", id)

	for {
		bts, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			log.Printf("read message error: %v", err)
			return
		}
		ProcessMessage(s.clients, client, bts)
	}
}
