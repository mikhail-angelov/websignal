package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoom(t *testing.T) {
	rooms := NewRooms()
	err := rooms.addRoom("test", "test")
	require.NoError(t, err)
	err = rooms.addRoom("test", "test")
	require.EqualError(t, err, "already exist")
}
