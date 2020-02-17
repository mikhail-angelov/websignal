package server

import "encoding/json"

const (
	textMessage                int = 0
	createRoomMessage              = 1
	joinRoomMessage                = 2
	leaveRoomMessage               = 3
	sdpMessage                     = 4
	candidateMessage               = 5
	getRoomsMessage                = 6
	roomIsCreatedMessage           = 7
	roomUpdateMessage              = 8
	startPeerConnectionMessage     = 9
	addFakeUser                    = 10
	removeFakeUser                 = 11
)

// Message (ws) fields
type Message struct {
	From string          `json:"from"`
	To   string          `json:"to"`
	Type int             `json:"type"`
	Data json.RawMessage `json:"data"`
}

// InputMessageData generic input/output data format
type InputMessageData map[string]string
