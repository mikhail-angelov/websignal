package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRoom(t *testing.T) {
	rooms := NewRoomService()
	room, err := rooms.CreateRoom("test")
	require.NoError(t, err)
	require.Equal(t, room.Owner, "test")
}

func TestRemoveRoom(t *testing.T) {
	rooms := NewRoomService()
	room, err := rooms.CreateRoom("test")

	err = rooms.RemoveRoom("test1", "test")
	require.EqualError(t, err, "does not exist")
	err = rooms.RemoveRoom(room.ID, "test2")
	require.EqualError(t, err, "test2 is not owner")
	err = rooms.RemoveRoom(room.ID, "test")
	require.NoError(t, err)
}

func TestJoinLeaveRoom(t *testing.T) {
	rooms := NewRoomService()
	room, err := rooms.CreateRoom("test")

	room, err = rooms.JoinToRoom(room.ID, "test2")
	users, _ := rooms.GetRoomUsers(room.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(users))
	assert.Equal(t, "test2", users[1])

	err = rooms.LeaveRoom(room.ID, "test2")
	users, _ = rooms.GetRoomUsers(room.ID)
	assert.Equal(t, 1, len(users))
}
