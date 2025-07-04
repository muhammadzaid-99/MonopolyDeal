package main

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/lxzan/gws"
)

type RoomManager struct {
	Rooms            map[string]*Room
	Mutex            sync.Mutex
	PlayerIDToRoomID map[string]string
	Cleanup          chan string
}

func NewRoomManager() *RoomManager {
	return &RoomManager{Rooms: make(map[string]*Room), PlayerIDToRoomID: make(map[string]string), Cleanup: make(chan string)}
}

func (rm *RoomManager) StartRoomCleanup() {
	fmt.Println("Room Cleaning Service Started")
	go func() {
		for roomID := range rm.Cleanup {
			rm.Mutex.Lock()
			if room, exists := rm.Rooms[roomID]; exists {
				delete(rm.Rooms, roomID)
				room.Players = nil
				room.PlayerOrder = nil
				room.Game = nil
				close(room.Events)
			}
			rm.Mutex.Unlock()
			fmt.Printf("Room %s cleaned up\n", roomID)
			fmt.Printf("Rooms count: %v\n", len(rm.Rooms))
		}
	}()
}

func (rm *RoomManager) NewRoom(roomID string) *Room {
	r := &Room{
		ID:      roomID,
		Players: make(map[string]*PlayerConn),
		Events:  make(chan RoomEvent, 10),
		Quit:    make(chan struct{}),
		Cleanup: rm.Cleanup,
	}
	rm.Rooms[roomID] = r
	go r.Run()
	return r
}

func (rm *RoomManager) UpdatePlayerConnection(playerID string, c *gws.Conn) {
	room := rm.GetPlayerRoom(playerID)
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()
	if room != nil {
		msg := WSMessage{
			Type:     "update-player-connection",
			PlayerID: playerID,
		}
		room.Events <- RoomEvent{c, msg, nil}
	}
}

func (rm *RoomManager) GetPlayerRoom(playerID string) *Room {
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()
	roomID, ok := rm.PlayerIDToRoomID[playerID]
	if !ok {
		return nil
	}
	return rm.Rooms[roomID]
}

func (rm *RoomManager) BroadcastAllGameState(playerID string) {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		// room.BroadcastAllGameState()
		gs := WSMessage{
			Type: "game-state",
		}
		room.Events <- RoomEvent{nil, gs, nil}
	}
}

func (rm *RoomManager) InformIPChange(playerID string) {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		msg := WSMessage{
			Type:     "broadcast-message",
			PlayerID: playerID,
			Message:  "IP Changed",
		}
		room.Events <- RoomEvent{nil, msg, nil}
		// room.BroadcastMessage(playerID, "IP Changed")
	}
}

func (rm *RoomManager) TryLeaveRoom(playerID string) (leftRoom bool) {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		reply := make(chan error)
		leave := WSMessage{
			Type:     "leave-room",
			PlayerID: playerID,
		}
		room.Events <- RoomEvent{nil, leave, reply}
		if err := <-reply; err != nil {
			fmt.Println(err)
			return false
		} else {
			rm.BroadcastAllGameState(playerID)
			delete(rm.PlayerIDToRoomID, playerID)
			return true
		}
	}
	return false
}

func (rm *RoomManager) InformDisconnect(playerID string) {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		// room.BroadcastMessage(playerID, "Disconnected")
		disc := WSMessage{
			Type:     "broadcast-message",
			PlayerID: playerID,
			Message:  "Disconnected",
		}
		room.Events <- RoomEvent{nil, disc, nil}

	}
}

func (rm *RoomManager) InformConnect(playerID string) {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		msg := WSMessage{
			Type:     "broadcast-message",
			PlayerID: playerID,
			Message:  "Connected",
		}
		room.Events <- RoomEvent{nil, msg, nil}
		// room.BroadcastMessage(playerID, "Connected")

		rm.BroadcastAllGameState(playerID)
		// room.BroadcastPlayerList()
	}
}

func (rm *RoomManager) RoomExists(roomID string) bool {
	_, ok := rm.Rooms[roomID]
	return ok
}

func (rm *RoomManager) GenerateRoomID() string {
	// Generate a unique room ID using a random UUID.
	for {
		code := fmt.Sprintf("%06d", 100000+int(uuid.New().ID()%900000))
		if !rm.RoomExists(code) {
			return code
		}
	}
}

func (rm *RoomManager) CreateJoinRoom(c *gws.Conn, msg WSMessage) (string, bool) {
	if id, ok := rm.PlayerIDToRoomID[msg.PlayerID]; ok {
		c.WriteMessage(gws.OpcodeText, fmt.Appendf(nil, `{"type":"already-in-a-room"}`))
		return id, false
	}
	roomID := rm.GenerateRoomID()
	rm.NewRoom(roomID)

	msg.Type = "join-room"
	return rm.JoinRoom(c, msg, roomID)
}

func (rm *RoomManager) JoinRoom(c *gws.Conn, msg WSMessage, roomID string) (string, bool) {
	if rid, ok := rm.PlayerIDToRoomID[msg.PlayerID]; ok {
		c.WriteMessage(gws.OpcodeText, fmt.Appendf(nil, `{"type":"already-in-a-room"}`))
		return rid, false
	}
	if msg.PlayerName == "" {
		SendMessage(c, map[string]string{
			"type":    "incomplete-info",
			"message": "Missing Player Name",
		})
		return "", false
	}
	if room, exists := rm.Rooms[roomID]; exists {
		reply := make(chan error)
		room.Events <- RoomEvent{c, msg, reply}
		if err := <-reply; err != nil {
			fmt.Println(err)
			return "", false
		} else {
			rm.PlayerIDToRoomID[msg.PlayerID] = roomID
			return roomID, true
		}
	}
	return "", false
}

func (rm *RoomManager) ChangeReadyState(playerID string, state bool) bool {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		return room.ChangeReadyState(playerID, state)
	}
	return false
}

func (rm *RoomManager) StartGame(playerID string) bool {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		return room.StartGame(playerID)
	}
	return false
}

func (rm *RoomManager) PlayCard(playerID string, cardID uint8, location PlayLocation) bool {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		return room.PlayCard(playerID, cardID, location)
	}
	return false
}

func (rm *RoomManager) ArrangeProperty(playerID string, cardID uint8, destination_pile_id string) bool {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		return room.ArrangeProperty(playerID, cardID, destination_pile_id)
	}
	return false
}

func (rm *RoomManager) DrawCards(playerID string) bool {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		return room.DrawCards(playerID)
	}
	return false
}

func (rm *RoomManager) DiscardCard(playerID string, cardID uint8) bool {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		return room.DiscardCard(playerID, cardID)
	}
	return false
}

func (rm *RoomManager) EndTurn(playerID string) bool {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		return room.EndTurn(playerID)
	}
	return false
}
