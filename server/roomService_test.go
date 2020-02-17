package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRoom(t *testing.T) {
	rooms := NewRoomService()
	user := User{ID: "test", PeerID: "test-peer", Name: "test"}
	room, err := rooms.CreateRoom(user)
	require.NoError(t, err)
	require.Equal(t, room.Owner, user.PeerID)
}

func TestRemoveRoom(t *testing.T) {
	rooms := NewRoomService()
	user := User{ID: "test", PeerID: "test-peer", Name: "test"}
	room, err := rooms.CreateRoom(user)

	err = rooms.RemoveRoom("test1", "test")
	require.EqualError(t, err, "does not exist")
	err = rooms.RemoveRoom(room.ID, "test2")
	require.EqualError(t, err, "test2 is not owner")
	err = rooms.RemoveRoom(room.ID, user.PeerID)
	require.NoError(t, err)
}

func TestJoinLeaveRoom(t *testing.T) {
	roomService := NewRoomService()
	user := User{ID: "test", PeerID: "test-peer", Name: "test"}
	room, err := roomService.CreateRoom(user)
	require.NoError(t, err)

	user2 := User{ID: "test2", PeerID: "test-peer2", Name: "test2"}
	room, err = roomService.JoinToRoom(room.ID, user2)
	require.NoError(t, err)
	rooms, err := roomService.GetUserRooms(user2.PeerID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(rooms))
	assert.Equal(t, user.PeerID, rooms[0].Owner)

	_, err = roomService.LeaveRoom(room.ID, user2.PeerID)
	require.NoError(t, err)
	rooms, _ = roomService.GetUserRooms(user2.PeerID)
	assert.Equal(t, 0, len(rooms))
}
