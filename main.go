package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/lxzan/gws"
)

func LoadEnvironmentVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func TestDBConnection() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close(context.Background())

	// Example query to test connection
	var version string
	if err := conn.QueryRow(context.Background(), "SELECT version()").Scan(&version); err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	log.Println("Supabase:: Connected to:", version)
}

func init() {

	go LoadEnvironmentVariables()
	go TestDBConnection()

}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local
	}
	fmt.Println("PORT for Server will be", port)

	h := &WSHandler{
		RoomMgr:        NewRoomManager(),
		PlayerIDToIP:   make(map[string]string),
		ConnToPlayerID: make(map[*gws.Conn]string),
	}
	server := gws.NewServer(h, nil)
	h.RoomMgr.StartRoomCleanup()

	log.Fatal(server.Run(":" + port))
}
