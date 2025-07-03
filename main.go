package main

import (
	"github.com/lxzan/gws"
)

func main() {
	server := gws.NewServer(&WSHandler{
		RoomMgr:        NewRoomManager(),
		PlayerIDToIP:   make(map[string]string),
		ConnToPlayerID: make(map[*gws.Conn]string),
	}, nil)

	server.Run(":4040")
}
