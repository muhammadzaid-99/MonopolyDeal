package game

import (
	c "cashdeal/models/cards"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrMaxPlayersReached   = errors.New("maximum number of players reached")
	ErrPlayerAlreadyExists = errors.New("player already exists in the game")
)

const (
	MaxPlayers           = 6
	MinPlayers           = 2
	MaxHandCards         = 7
	RequiredPropertySets = 3
	CardsDrawnPerTurn    = 2
	CardsDrawnOnRunOut   = 5
	CardsDrawnPassGo     = 2
	MaxCardsPlayable     = 3
	MaxPilesAllowed      = 20
	HouseRent            = 3
	HotelRent            = 4 // including house
)

type Game struct {
	Deck               []c.Card
	CurrentTurn        *Player2
	CardPlaysRemaining uint8
	TurnCardsDrawn     bool
	DiscardPile        []c.Card
	TopCard            c.Card
	Mutex              sync.Mutex
	Messages           []string
}

func NewGame() *Game {
	return &Game{
		Deck:        c.GetNewDeck(),
		CurrentTurn: nil,
		DiscardPile: []c.Card{},
	}
}

func (g *Game) PushNewMessage(message string) {
	if len(g.Messages) >= 32 {
		g.Messages = append([]string{message}, g.Messages[:len(g.Messages)-1]...)
	} else {
		g.Messages = append([]string{message}, g.Messages...)
	}
}

func (g *Game) MoveToDiscardPile(card c.Card) {
	g.Mutex.Lock()
	g.DiscardPile = append(g.DiscardPile, card)
	g.TopCard = card
	g.Mutex.Unlock()
}

func (g *Game) DrawCard() c.Card {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()
	if len(g.Deck) == 0 {
		fmt.Println("Shuffled discard pile for deck")
		if len(g.DiscardPile) > 1 {
			c.ShuffleDeck(g.DiscardPile[:len(g.DiscardPile)-1])
			g.Deck = append([]c.Card(nil), g.DiscardPile[:len(g.DiscardPile)-1]...)
			g.DiscardPile = g.DiscardPile[len(g.DiscardPile)-1:]
		} else {
			// No cards left to draw
			return nil
		}
	}
	card := g.Deck[0]
	g.Deck = g.Deck[1:]
	return card
}

func (g *Game) NewPlayer(playerID string, playerName string) *Player2 {
	// PropertyCards := make(map[c.PropertyColor][]c.Card)
	// for _, color := range c.GetAllColors() {
	// 	PropertyCards[color] = nil
	// }

	return &Player2{
		ID:            playerID,
		Name:          playerName,
		HandCards:     make(map[uint8]c.Card),
		ActionMessage: make(map[string]string),
		PropertyPiles: make(map[string]*PropertyPile),
	}
	// return &Player{
	// 	ID:                 playerID,
	// 	Name:               playerName,
	// 	HandCards:          make(map[uint8]c.Card),
	// 	PropertyCards:      PropertyCards,
	// 	PropertyRents:      make(map[c.PropertyColor]c.MoneyValue),
	// 	PropertyCompletion: make(map[c.PropertyColor]uint8),
	// 	ActionMessage:      make(map[string]string),
	// }
}
