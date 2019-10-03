package server

const (
	textMessage int = iota
	createRoomMessage
	joinRoomMessage
	leaveRoomMessage
	sdpMessage
	candidateMessage
)

// Message (ws) fields
type Message struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type int    `json:"type"`
	Data string `json:"data"`
}
