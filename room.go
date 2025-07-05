package main

import (
	c "cashdeal/models/cards"
	g "cashdeal/models/game"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/lxzan/gws"
)

type PlayerConn struct {
	g.Player2
	Conn *gws.Conn
}

type PlayLocation string

const (
	DiscardPile  PlayLocation = "DiscardPile"
	BankPile     PlayLocation = "BankPile"
	PropertyPile PlayLocation = "PropertyPile"
)

type RoomEvent struct {
	c     *gws.Conn
	msg   WSMessage
	reply chan error
}

type Room struct {
	Players       map[string]*PlayerConn
	Game          *g.Game
	Mutex         sync.Mutex
	ID            string
	IsGameStarted bool
	PlayerOrder   []string
	TurnIndex     int
	PendingAction PendingAction
	Events        chan RoomEvent
	Quit          chan struct{}
	DeleteAt      *time.Timer
	Cleanup       chan string
}

func (r *Room) TryDelete() {
	// either we have no players or players who are not Connected
	for _, p := range r.Players {
		if p.Connected {
			return
		}
	}
	duration := time.Minute
	if len(r.Players) == 0 {
		duration *= 2
	} else {
		duration *= 20
	}
	fmt.Printf("The room %s will be deleted after %v at %s.\n", r.ID, duration, time.Now().Add(duration).Format("2006-01-02 15:04:05"))
	r.DeleteAt = time.AfterFunc(duration, func() {

		close(r.Quit)
	})
}

func (r *Room) JoinRoom(playerID string, playerName string, c *gws.Conn) error {

	if r.IsGameStarted {
		c.WriteMessage(gws.OpcodeText, fmt.Appendf(nil, `{"type":"game-already-started"}`))
		return errors.New("error: game started already")
	}
	if len(r.Players) >= g.MaxPlayers {
		c.WriteMessage(gws.OpcodeText, fmt.Appendf(nil, `{"type":"max-players-joined"}`))
		return errors.New("error: room is full")
	}
	// fmt.Println("Now adding player")
	r.AddPlayer(playerID, playerName, c)
	r.ResetDeletion()
	return nil

}

func (r *Room) ResetDeletion() {
	if r.DeleteAt != nil {
		fmt.Println("Room deletion cancelled: ", r.ID)
		r.DeleteAt.Stop()
		r.DeleteAt = nil
	}
}

func (r *Room) LeaveRoom(playerID string) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	defer r.TryDelete()
	if p, ok := r.Players[playerID]; ok {
		p.Connected = false // could not leave but we know player disconnected
		if r.IsGameStarted {
			return errors.New("cannot-leave-while-game-is-started")
		}
		delete(r.Players, playerID)
		return nil
	}
	return errors.New("could-not-leave-room")
}

func (r *Room) UpdatePlayerConnection(c *gws.Conn, playerID string) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if player, ok := r.Players[playerID]; ok {
		player.Conn = c
		player.Connected = true // we know player is now connected
		r.ResetDeletion()
	}
}

func (r *Room) BroadcastMessage(playerID string, message string) {
	for _, p := range r.Players {
		SendMessage(p.Conn, map[string]string{"type": "broadcast", "message": message, "player_id": playerID})
	}
}

func (r *Room) GetPlayerListForBroadcast() []byte {
	players := make([]map[string]any, 0, len(r.Players))
	for _, p := range r.Players {
		playerInfo := map[string]any{
			"ID":            p.ID,
			"Name":          p.Name,
			"IsReady":       p.IsReady,
			"BankCards":     p.BankCards,
			"HandCount":     len(p.HandCards),
			"PropertyPiles": p.PropertyPiles,
			"LooseCards":    p.LooseCards,
			"BankSum":       p.GetBankSum(),
			// "PropertyCards":      p.PropertyCards,
			// "PropertyRents":      p.PropertyRents,
			// "PropertyCompletion": p.PropertyCompletion,
		}

		players = append(players, playerInfo)
	}

	msg := map[string]any{
		"type":    "player-list",
		"players": players,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return nil
	}
	return b
}

func (r *Room) GetGameInfoForBroadcast() []byte {
	if !r.IsGameStarted {
		// game not being played
		return nil
	}
	game := r.Game
	gameState := map[string]any{
		"type":                   "game-info",
		"current_turn_player_id": game.CurrentTurn.ID,
		"top_card":               game.TopCard,
		"plays_remaining":        game.CardPlaysRemaining,
		"turn_cards_drawn":       game.TurnCardsDrawn,
		"messages":               game.Messages,
	}

	gi, err := json.Marshal(gameState)
	if err != nil {
		return nil
	}
	return gi
}

func (r *Room) BroadcastPlayerList() {
	b := r.GetPlayerListForBroadcast()
	for _, p := range r.Players {
		p.Conn.WriteMessage(gws.OpcodeText, b)
	}
}

func (r *Room) BroadcastGameInfo() {
	gi := r.GetGameInfoForBroadcast()
	if gi != nil {
		for _, p := range r.Players {
			p.Conn.WriteMessage(gws.OpcodeText, gi)
		}
	}
}

func (r *Room) BroadcastPlayersHands() {
	for _, p := range r.Players {
		r.NotifyPlayerHand(p)
	}
}

func (r *Room) BroadcastAllGameState() {
	b := r.GetPlayerListForBroadcast()
	gi := r.GetGameInfoForBroadcast()
	for _, p := range r.Players {
		r.NotifyPlayerHand(p)
		if b != nil {
			p.Conn.WriteMessage(gws.OpcodeText, b)
		}
		if gi != nil {
			p.Conn.WriteMessage(gws.OpcodeText, gi)
		}
		SendMessage(p.Conn, p.ActionMessage)
	}
}

func (r *Room) SetPlayerActionMessage(player *PlayerConn, msg map[string]string) {
	player.ActionMessage = msg
	SendMessage(player.Conn, msg)
}

func (r *Room) IsEveryoneReady() bool {
	for _, p := range r.Players {
		if !p.IsReady {
			return false
		}
	}
	return true
}

func (r *Room) AddPlayer(playerID string, playerName string, c *gws.Conn) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	_, exists := r.Players[playerID]
	if !exists {
		for _, p := range r.Players {
			if p.Name == playerName {
				playerName = fmt.Sprintf("%s%d", playerName, rand.Intn(10))
			}
		}
		r.Players[playerID] = &PlayerConn{
			Conn:    c,
			Player2: *r.Game.NewPlayer(playerID, playerName),
		}
		r.BroadcastMessage(playerID, fmt.Sprintf("New Player Joined: %s", playerName))
		r.BroadcastPlayerList()
	} else {
		c.WriteMessage(gws.OpcodeText, fmt.Appendf(nil, `{"type":"already-in-room"}`))
	}
}

func (r *Room) ChangeReadyState(playerID string, state bool) bool {
	r.Mutex.Lock()
	r.Players[playerID].IsReady = state
	r.Mutex.Unlock()
	r.BroadcastPlayerList()
	return state
}

func (r *Room) NotifyPlayerHand(player *PlayerConn) {
	handMsg := map[string]any{
		"type":      "hand-cards",
		"handCards": player.HandCards,
	}
	// fmt.Println(handMsg)
	b, _ := json.Marshal(handMsg)
	player.Conn.WriteMessage(gws.OpcodeText, b)
}

func (r *Room) DealCards() {
	// Deal initial cards to each player at the start of the game
	if r.Game == nil {
		return
	}
	for _, p := range r.Players {
		// Draw 5 cards for each player (or whatever the starting hand is)
		for range g.CardsDrawnOnRunOut {
			card := r.Game.DrawCard()
			if card != nil {
				p.AddToHand(card)
				// fmt.Println(card.GetID(), "-", p.GetHandCount(), "-", len(r.Game.Deck))
			} else {
				fmt.Println("A nil card was drawn")
			}
		}

		// Notify the player of their hand
		r.NotifyPlayerHand(p)
	}

}

func (r *Room) StartGame(playerID string) bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if r.IsGameStarted {
		// game already started
		return false
	}

	if len(r.Players) < g.MinPlayers {
		return false
	}

	if r.IsEveryoneReady() {
		// first lock the room
		r.IsGameStarted = true
		r.Game = g.NewGame()
		r.DealCards()
		r.SetPlayerOrder()
		r.BroadcastAllGameState()

		return true
	}
	return false
}

func (r *Room) SetPlayerOrder() {
	r.PlayerOrder = slices.Collect(maps.Keys(r.Players))
	r.TurnIndex = rand.Intn(len(r.Players))
	r.Game.CurrentTurn = &r.Players[r.PlayerOrder[r.TurnIndex]].Player2
	r.Game.CardPlaysRemaining = g.MaxCardsPlayable
	r.Game.TurnCardsDrawn = false
}

func (r *Room) NextTurn() {
	r.TurnIndex = (r.TurnIndex + 1) % len(r.PlayerOrder)
	r.Game.CurrentTurn = &r.Players[r.PlayerOrder[r.TurnIndex]].Player2
	r.Game.TurnCardsDrawn = false
	r.Game.CardPlaysRemaining = g.MaxCardsPlayable
	r.Game.PushNewMessage(fmt.Sprintf("%s's turn now", r.Game.CurrentTurn.Name))
}

func (r *Room) CheckPlayerTurn(player *PlayerConn) bool {
	if &player.Player2 != r.Game.CurrentTurn {
		SendMessage(player.Conn, map[string]string{
			"type": "not-your-turn",
		})
		return false
	}
	return true
}

func (r *Room) HasPlayRemaining() bool {
	return r.Game.CardPlaysRemaining != 0
}

func (r *Room) CreatePile(playerID string) bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if !r.IsGameStarted {
		// game not being played
		return false
	}

	// in your turn only
	player := r.Players[playerID]
	if !r.CheckPlayerTurn(player) {
		return false
	}

	if len(player.PropertyPiles) < g.MaxPilesAllowed {
		player.NewPile()
		r.BroadcastAllGameState()
		return true
	}
	return false
}

func (r *Room) ChangePileColor(playerID string, pile_id string, color c.PropertyColor) bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if !r.IsGameStarted {
		// game not being played
		return false
	}

	// in your turn only
	player := r.Players[playerID]
	if !r.CheckPlayerTurn(player) {
		return false
	}

	player.ChangePileColor(pile_id, color)
	r.BroadcastAllGameState()
	return true
}

func (r *Room) PlayCard(playerID string, cardID uint8, location PlayLocation) bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if !r.IsGameStarted {
		// game not being played
		return false
	}

	//in your turn only
	player := r.Players[playerID]
	if !r.CheckPlayerTurn(player) {
		return false
	}

	// only after drawing cards
	if !r.Game.TurnCardsDrawn {
		SendMessage(player.Conn, map[string]string{
			"type": "draw-cards-first",
		})
		return false
	}

	// only if plays remaining
	if !r.HasPlayRemaining() {
		SendMessage(player.Conn, map[string]string{
			"type": "no-plays-remaining",
		})
		return false
	}

	card, ok := player.GetFromHand(cardID)
	if ok {
		fmt.Println("Card Playing ", card.GetID())
		switch location {
		case DiscardPile:
			fmt.Println("In Discard Pile")
			fmt.Println("Card type is ", card.GetType())
			if card.GetType() != c.CardTypeAction && card.GetType() != c.CardTypeRent {
				player.AddToHand(card) // rollback
				fmt.Println("Invalid Card Type for Playing")
				return false
			}
			if !r.HandleActionCardPlay(card) {
				player.AddToHand(card) // card unplayable
				fmt.Println("Unplayable")
				return false
			}
			// card was playable
			fmt.Println("Played")
			r.Game.MoveToDiscardPile(card)
			r.Game.PushNewMessage(fmt.Sprintf("%s played %s", player.Name, card.GetName()))
		case BankPile:
			// panic("BANK PANIC!!!")
			fmt.Println("In Bank Pile")
			if !player.AddToBank(card) {
				player.AddToHand(card) // rollback
				return false
			}
			r.Game.PushNewMessage(fmt.Sprintf("%s moved %s card of $%v to bank", player.Name, card.GetName(), card.GetMoneyValue()))
		case PropertyPile:
			fmt.Println("In Property Pile")
			// "" will add to loose cards
			if !player.AddToProperty("", card) {
				player.AddToHand(card) // rollback
				return false
			}
			r.Game.PushNewMessage(fmt.Sprintf("%s moved %s card to properties", player.Name, card.GetName()))
		default:
			return false
		}
		r.Game.CardPlaysRemaining--
		r.BroadcastAllGameState()
		return true
	}

	return false
}

func (r *Room) ArrangeProperty(playerID string, cardID uint8, destination_pile_id string) bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if !r.IsGameStarted {
		// game not being played
		return false
	}

	// in your turn only
	player := r.Players[playerID]
	if !r.CheckPlayerTurn(player) {
		return false
	}

	// as many times as want
	if card, ok, _, commit := player.GetProperty(cardID); ok {
		// pile_id is not sent so it will be "", will add to loose cards
		if player.AddToProperty(destination_pile_id, card) {
			// commit will remove from original source
			commit()
			r.Game.PushNewMessage(fmt.Sprintf("%s arranged their %s property", player.Name, card.GetName()))
			r.BroadcastAllGameState()
			r.CheckWin(&player.Player2)
			return true
		}
	}

	return false
}

func (r *Room) DrawCards(playerID string) bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if !r.IsGameStarted {
		// game not being played
		return false
	}

	// in your turn
	player := r.Players[playerID]
	if !r.CheckPlayerTurn(player) {
		return false
	}

	// if not drawn, only then
	if r.Game.TurnCardsDrawn {
		SendMessage(player.Conn, map[string]string{
			"type": "cards-already-drawn",
		})
		return false
	}

	// for empty hand draw 5
	if player.IsHandEmpty() {
		for range g.CardsDrawnOnRunOut {
			card := r.Game.DrawCard()
			if card != nil {
				player.AddToHand(card)
			}
		}
		r.Game.PushNewMessage(fmt.Sprintf("%s has drawn %v cards", player.Name, g.CardsDrawnOnRunOut))
	} else {
		// draw 2 else
		for range g.CardsDrawnPerTurn {
			card := r.Game.DrawCard()
			if card != nil {
				player.AddToHand(card)
			}
		}
		r.Game.PushNewMessage(fmt.Sprintf("%s has drawn %v cards", player.Name, g.CardsDrawnPerTurn))
	}

	r.Game.TurnCardsDrawn = true
	r.BroadcastAllGameState()
	return true

}

func (r *Room) DiscardCard(playerID string, cardID uint8) bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if !r.IsGameStarted {
		// game not being played
		return false
	}

	// discard in your turn
	player := r.Players[playerID]
	if !r.CheckPlayerTurn(player) {
		return false
	}

	// discard after playing only
	if r.HasPlayRemaining() {
		return false
	}

	if player.GetHandCount() > g.MaxHandCards {
		if card, ok := player.GetFromHand(cardID); ok {
			r.Game.MoveToDiscardPile(card)
			r.Game.PushNewMessage(fmt.Sprintf("%s has discarded %s card", player.Name, card.GetName()))
			r.BroadcastAllGameState()
			return true
		}
	}
	return false
}

func (r *Room) EndTurn(playerID string) bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if !r.IsGameStarted {
		// game not being played
		return false
	}

	player := r.Players[playerID]
	if !r.CheckPlayerTurn(player) {
		return false
	}

	// only after drawing cards
	if !r.Game.TurnCardsDrawn {
		SendMessage(player.Conn, map[string]string{
			"type": "draw-cards-first",
		})
		return false
	}

	// player can end turn even if plays remaining

	// after discarding extra cards
	if player.GetHandCount() > g.MaxHandCards {
		SendMessage(player.Conn, map[string]string{
			"type": "discard-extra-cards",
		})
		return false
	}

	r.PendingAction = nil
	player.DeleteEmptyPiles()
	r.NextTurn()
	r.BroadcastAllGameState()
	return true
}

func (r *Room) CheckWin(player *g.Player2) {
	if player.HasRequiredPropertySets() {
		r.Game.PushNewMessage(fmt.Sprintf("%s wins", player.Name))
		r.BroadcastAllGameState()
		r.IsGameStarted = false
		r.Game = nil
		for _, p := range r.Players {
			p.HandCards = nil
			p.PropertyPiles = nil
			p.ActionMessage = nil
			p.LooseCards = nil
			p.IsReady = false
		}
	}
}
