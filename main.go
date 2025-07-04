package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lxzan/gws"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local
	}
	fmt.Println("PORT is ", port)

	h := &WSHandler{
		RoomMgr:        NewRoomManager(),
		PlayerIDToIP:   make(map[string]string),
		ConnToPlayerID: make(map[*gws.Conn]string),
	}
	server := gws.NewServer(h, nil)
	h.RoomMgr.StartRoomCleanup()

	log.Fatal(server.Run(":" + port))
}
