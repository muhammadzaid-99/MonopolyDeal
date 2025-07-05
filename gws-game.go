package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/lxzan/gws"
)

func (h *WSHandler) GameMessageHandler(c *gws.Conn, msg WSMessage) {

	room := h.RoomMgr.GetPlayerRoom(msg.PlayerID)
	if room == nil {
		return
	}
	reply := make(chan error)
	room.Events <- RoomEvent{c, msg, reply}
}

func (r *Room) Run() {
	defer func() {
		if rec := recover(); rec != nil {
			stack := debug.Stack()
			roomID := r.ID
			log.Printf("Room %s crashed: %v\n%s\n", roomID, rec, stack)

			// Notify manager
			r.Cleanup <- r.ID

			go func() {
				dbConn, dberr := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
				if dberr != nil {
					log.Printf("Failed to connect to the database: %v", dberr)
					return
				}
				defer dbConn.Close(context.Background())

				_, err := dbConn.Exec(context.Background(), `
				INSERT INTO crash_logs (room_id, error_message, stack_trace, created_at)
				VALUES ($1, $2, $3, NOW())
			`, roomID, fmt.Sprintf("%v", rec), stack)

				if err != nil {
					log.Printf("Failed to insert crash log: %v", err)
				}
			}()
		} else {
			r.Cleanup <- r.ID
		}
	}()

	for {
		select {
		case event := <-r.Events:
			r.handleEvent(event)
		case <-r.Quit:
			return
		}
	}
}

func (r *Room) handleEvent(event RoomEvent) {
	msg := event.msg
	c := event.c
	fmt.Println("New Room Event :: RoomID:", r.ID, "Type:", msg.Type)

	switch msg.Type {
	case "join-room":
		err := r.JoinRoom(msg.PlayerID, msg.PlayerName, c)
		SendMessage(c, map[string]string{
			"type":        msg.Type,
			"room_id":     msg.RoomID,
			"join_status": strconv.FormatBool(err == nil),
		})
		event.reply <- err
	case "game-state":
		r.BroadcastAllGameState()
	case "change-ready-state":
		readyState := msg.ReadyState
		state := r.ChangeReadyState(msg.PlayerID, readyState)
		SendMessage(c, map[string]string{
			"type":        msg.Type,
			"ready_state": strconv.FormatBool(state),
		})
	case "start-game":
		ok := r.StartGame(msg.PlayerID)
		SendMessage(c, map[string]string{
			"type":       msg.Type,
			"game_start": strconv.FormatBool(ok),
			"message":    "If game_start is false, either not everyone is ready or game already started.",
		})
	case "leave-room": // only for room manager to access
		err := r.LeaveRoom(msg.PlayerID)
		event.reply <- err
	case "broadcast-message":
		r.BroadcastMessage(msg.PlayerID, msg.Message)
	case "update-player-connection":
		r.UpdatePlayerConnection(c, msg.PlayerID)
	}

	if !r.IsGameStarted {
		// game not being played
		return
	}

	if r.PendingAction != nil {
		// it should be noted that we are not acquiring lock because
		// it is guaranteed that there is no other path to move
		// acquiring lock will cause deadlocks here
		r.Mutex.Lock()
		if player, ok := r.Players[msg.PlayerID]; ok {
			r.Mutex.Unlock()
			r.PendingAction.Resolve(r, player, msg)
		} else {
			r.Mutex.Unlock()
			return
		}
		r.BroadcastAllGameState()

	} else {
		switch msg.Type {
		case "play-card":
			fmt.Println(msg.CardID)
			r.PlayCard(msg.PlayerID, msg.CardID, msg.PlayLocation)
		case "arrange-property":
			fmt.Println(msg.CardID, msg.PropertyArrangeDestPileID)
			r.ArrangeProperty(msg.PlayerID, msg.CardID, msg.PropertyArrangeDestPileID)
		case "end-turn":
			r.EndTurn(msg.PlayerID)
		case "draw-cards":
			r.DrawCards(msg.PlayerID)
		case "discard-card":
			r.DiscardCard(msg.PlayerID, msg.CardID)
		case "create-pile":
			r.CreatePile(msg.PlayerID)
		case "change-pile-color":
			r.ChangePileColor(msg.PlayerID, msg.PropertyPileID, msg.PropertyColor)
		}
	}
}
