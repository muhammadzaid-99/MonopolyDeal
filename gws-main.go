package main

import (
	"cashdeal/models/cards"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/lxzan/gws"
)

type WSHandler struct {
	gws.BuiltinEventHandler
	RoomMgr        *RoomManager
	PlayerIDToIP   map[string]string    // to prevent spoofing
	ConnToPlayerID map[*gws.Conn]string // to know who disconnected, check reconnection
	// actually we could provide client with Conn instead of player id, but the data structure would be same size
	// i.e. there will be a conn to conn map, key: conn reassigned, value: conn initially assigned
	// player ids seem more intuitive and less confusing, as one unique identifier must not be changed
	// in other case, client will have to store changing conn, now they save player id once
	Mutex sync.RWMutex
}

type WSMessage struct {
	Type                      string              `json:"type"`
	PlayerID                  string              `json:"player_id,omitempty"`
	RoomID                    string              `json:"room_id,omitempty"`
	PlayerName                string              `json:"player_name,omitempty"`
	ReadyState                bool                `json:"ready_state,omitempty"`
	CardID                    uint8               `json:"card_id,omitempty"`
	PlayLocation              PlayLocation        `json:"play_location,omitempty"`
	PropertyArrangeDestPileID string              `json:"property_arrange_dest_pile_id,omitempty"`
	TargetPlayerID            string              `json:"target_player_id,omitempty"`
	PropertyPileID            string              `json:"property_pile_id,omitempty"`
	PropertyColor             cards.PropertyColor `json:"property_color,omitempty"`
	// PropertyArrangeDestColor cards.PropertyColor `json:"property_arrange_dest_color,omitempty"`
}

func SendMessage(c *gws.Conn, message map[string]string) {
	data, _ := json.Marshal(message)
	c.WriteMessage(gws.OpcodeText, data)
}

func (h *WSHandler) GeneratePlayerID() string {
	// Generate a unique player ID using a random UUID.
	return fmt.Sprintf("player-%s", uuid.NewString())
}

func (h *WSHandler) OnOpen(c *gws.Conn) {
	ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
	fmt.Println("WebSocket connection opened:", ip)
	c.WriteMessage(gws.OpcodeText, fmt.Appendf(nil, `{"type":"connected"}`))
}

func (h *WSHandler) ListConnection(c *gws.Conn, playerID string) {
	h.Mutex.Lock()
	h.ConnToPlayerID[c] = playerID
	h.Mutex.Unlock()
}

func (h *WSHandler) ListIP(playerID string, ip string) {
	h.Mutex.Lock()
	h.PlayerIDToIP[playerID] = ip
	h.Mutex.Unlock()
}

func (h *WSHandler) DeleteConnection(c *gws.Conn) {
	h.Mutex.Lock()
	delete(h.ConnToPlayerID, c)
	h.Mutex.Unlock()
}

func (h *WSHandler) GetIPFromConn(c *gws.Conn) string {
	ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
	return ip
}

func (h *WSHandler) OnMessage(c *gws.Conn, message *gws.Message) {
	var msg WSMessage
	if err := json.Unmarshal(message.Bytes(), &msg); err != nil {
		fmt.Println("invalid JSON:", err)
		return
	}
	playerID := msg.PlayerID
	fmt.Printf("A message came with player ID: %s and card ID: %v\n", playerID, msg.CardID)

	// First Message from New Player (Invalid Player ID passed)
	// whether player ID is empty or curated one, they cannot pass from here
	// it is checked that there is no issue if 2+ players join from same IP
	if _, ok := h.PlayerIDToIP[playerID]; !ok {
		playerID := h.GeneratePlayerID()
		h.Mutex.Lock()
		h.PlayerIDToIP[playerID] = h.GetIPFromConn(c)
		h.ConnToPlayerID[c] = playerID
		h.Mutex.Unlock()
		c.WriteMessage(gws.OpcodeText, fmt.Appendf(nil, `{"type":"player_id_assigned", "player_id":"%s"}`, playerID))
		fmt.Printf("Conn %p is for Player %s\n", c, playerID)
		return
	}

	// ###
	// at this point, it is guaranteed that a valid playerID exists
	// ###

	// If Player has Reconnected
	if h.ConnToPlayerID[c] != playerID {
		h.ListConnection(c, playerID)
		h.RoomMgr.UpdatePlayerConnection(playerID, c)
		h.RoomMgr.InformConnect(playerID)

		// If Player IP Changed
		// since IP change is not possible (i guess) without reconnection,
		// that's why we only check it on reconnection
		ip := h.GetIPFromConn(c)
		if h.PlayerIDToIP[playerID] != ip {
			h.ListIP(playerID, ip)
			h.RoomMgr.InformIPChange(playerID)
		}
	}

	switch msg.Type {

	case "create-room":
		roomID, joined := h.RoomMgr.CreateJoinRoom(c, msg)
		SendMessage(c, map[string]string{
			"type":    msg.Type,
			"room_id": roomID,
			"joined":  strconv.FormatBool(joined),
		})

	case "join-room":
		if msg.RoomID == "" {
			SendMessage(c, map[string]string{
				"type":    "incomplete-info",
				"message": "Missing Room ID",
			})
			return
		}
		roomID, joined := h.RoomMgr.JoinRoom(c, msg, msg.RoomID)
		SendMessage(c, map[string]string{
			"type":        msg.Type,
			"room_id":     roomID,
			"join_status": strconv.FormatBool(joined),
		})

	// case "game-state":
	// 	h.RoomMgr.BroadcastAllGameState(playerID)
	// case "change-ready-state":
	// 	readyState := msg.ReadyState
	// 	state := h.RoomMgr.ChangeReadyState(msg.PlayerID, readyState)
	// 	SendMessage(c, map[string]string{
	// 		"type":        msg.Type,
	// 		"ready_state": strconv.FormatBool(state),
	// 	})
	// case "start-game":
	// 	ok := h.RoomMgr.StartGame(msg.PlayerID)
	// 	SendMessage(c, map[string]string{
	// 		"type":       msg.Type,
	// 		"game_start": strconv.FormatBool(ok),
	// 		"message":    "If game_start is false, either not everyone is ready or game already started.",
	// 	})

	default:
		// further logic maintained by handler
		h.GameMessageHandler(c, msg)
	}

}

func (h *WSHandler) OnClose(c *gws.Conn, err error) {
	h.Mutex.Lock()
	playerID := h.ConnToPlayerID[c]
	h.Mutex.Unlock()
	h.RoomMgr.InformDisconnect(playerID)
	h.DeleteConnection(c)
	fmt.Printf("Player %s disconnected\n", playerID)
}
