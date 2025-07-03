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
}

func NewRoomManager() *RoomManager {
	return &RoomManager{Rooms: make(map[string]*Room), PlayerIDToRoomID: make(map[string]string)}
}

func (rm *RoomManager) NewRoom(roomID string) *Room {
	r := &Room{
		ID:      roomID,
		Players: make(map[string]*PlayerConn),
		Events:  make(chan RoomEvent, 10),
		Quit:    make(chan struct{}),
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
		player, ok := room.Players[playerID]
		if ok {
			player.Conn = c
		}
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
		room.BroadcastAllGameState()
	}
}

func (rm *RoomManager) InformIPChange(playerID string) {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		room.BroadcastMessage(playerID, "IP Changed")
	}
}

func (rm *RoomManager) InformDisconnect(playerID string) {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		room.BroadcastMessage(playerID, "Disconnected")
	}
}

func (rm *RoomManager) InformConnect(playerID string) {
	room := rm.GetPlayerRoom(playerID)
	if room != nil {
		room.BroadcastMessage(playerID, "Connected")
		room.BroadcastPlayerList()
	}
}

func (rm *RoomManager) RoomExists(roomID string) bool {
	_, ok := rm.Rooms[roomID]
	return ok
}

func (rm *RoomManager) GenerateRoomID() string {
	// Generate a unique room ID using a random UUID.
	for {
		code := fmt.Sprintf("%06d", uuid.New().ID()%1000000)
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

// func (rm *RoomManager) JoinRoom(playerID string, playerName string, roomID string, c *gws.Conn) (string, bool) {
// 	if rid, ok := rm.PlayerIDToRoomID[playerID]; ok {
// 		c.WriteMessage(gws.OpcodeText, fmt.Appendf(nil, `{"type":"already-in-a-room"}`))
// 		return rid, false
// 	}
// 	if room, exists := rm.Rooms[roomID]; exists {
// 		if room.IsGameStarted {
// 			c.WriteMessage(gws.OpcodeText, fmt.Appendf(nil, `{"type":"game-already-started"}`))
// 			return "", false
// 		}
// 		room.AddPlayer(playerID, playerName, c)
// 		rm.PlayerIDToRoomID[playerID] = roomID
// 		return roomID, true
// 	}
// 	return "", false
// }

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

// func (rm *RoomManager) ArrangeProperty(playerID string, cardID uint8, destination cards.PropertyColor) bool {
// 	room := rm.GetPlayerRoom(playerID)
// 	if room != nil {
// 		return room.ArrangeProperty(playerID, cardID, destination)
// 	}
// 	return false
// }

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
