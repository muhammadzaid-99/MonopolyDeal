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

	server := gws.NewServer(&WSHandler{
		RoomMgr:        NewRoomManager(),
		PlayerIDToIP:   make(map[string]string),
		ConnToPlayerID: make(map[*gws.Conn]string),
	}, nil)

	log.Fatal(server.Run(":" + port))
}
