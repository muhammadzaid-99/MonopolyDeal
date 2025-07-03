package main

import (
	"fmt"

	"github.com/lxzan/gws"
)

func (h *WSHandler) GameMessageHandler(c *gws.Conn, msg WSMessage) {

	room := h.RoomMgr.GetPlayerRoom(msg.PlayerID)
	if room == nil {
		return
	}

	if !room.IsGameStarted {
		// game not being played
		return
	}

	if room.PendingAction != nil {
		// it should be noted that we are not acquiring lock because
		// it is guaranteed that there is no other path to move
		// acquiring lock will cause deadlocks here
		room.PendingAction.Resolve(room, room.Players[msg.PlayerID], msg)
		room.BroadcastAllGameState()

	} else {
		switch msg.Type {
		case "play-card":
			fmt.Println(msg.CardID)
			room.PlayCard(msg.PlayerID, msg.CardID, msg.PlayLocation)
		case "arrange-property":
			fmt.Println(msg.CardID, msg.PropertyArrangeDestPileID)
			room.ArrangeProperty(msg.PlayerID, msg.CardID, msg.PropertyArrangeDestPileID)
		case "end-turn":
			room.EndTurn(msg.PlayerID)
		case "draw-cards":
			room.DrawCards(msg.PlayerID)
		case "discard-card":
			room.DiscardCard(msg.PlayerID, msg.CardID)
		case "create-pile":
			room.CreatePile(msg.PlayerID)
		case "change-pile-color":
			room.ChangePileColor(msg.PlayerID, msg.PropertyPileID, msg.PropertyColor)
		}
	}
}
