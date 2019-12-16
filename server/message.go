package server

const (
	textMessage          int = 0
	createRoomMessage        = 1
	joinRoomMessage          = 2
	leaveRoomMessage         = 3
	sdpMessage               = 4
	candidateMessage         = 5
	getRoomsMessage          = 6
	roomIsCreatedMessage     = 7
	roomToJoinMessage        = 8
)

// Message (ws) fields
type Message struct {
	From string                 `json:"from"`
	To   string                 `json:"to"`
	Type int                    `json:"type"`
	Data map[string]interface{} `json:"data"`
}
