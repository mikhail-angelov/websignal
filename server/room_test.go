package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRoom(t *testing.T) {
	rooms := NewRoomService()
	err := rooms.CreateRoom("test", "test")
	require.NoError(t, err)
	err = rooms.CreateRoom("test", "test")
	require.EqualError(t, err, "already exist")
}

func TestRemoveRoom(t *testing.T) {
	rooms := NewRoomService()
	err := rooms.CreateRoom("test", "test")

	err = rooms.RemoveRoom("test1", "test")
	require.EqualError(t, err, "does not exist")
	err = rooms.RemoveRoom("test", "test2")
	require.EqualError(t, err, "is not owner")
	err = rooms.RemoveRoom("test", "test")
	require.NoError(t, err)
}

func TestJoinLeaveRoom(t *testing.T) {
	rooms := NewRoomService()
	err := rooms.CreateRoom("test", "test")

	err = rooms.JoinToRoom("test", "test2")
	users, _ := rooms.GetRoomUsers("test")
	require.NoError(t, err)
	assert.Equal(t, 2, len(users))
	assert.Equal(t, "test2", users[1])

	err = rooms.LeaveRoom("test", "test2")
	users, _ = rooms.GetRoomUsers("test")
	assert.Equal(t, 1, len(users))
}
