package cards

import (
	"math/rand"
)

func GetNewDeck() []Card {
	// max 256 cards can be added
	var deck []Card

	// Money Cards (1-20, 20)
	deck = append(deck, GetMoneyCardDeck()...)

	// Action Cards (21-54, 34)
	deck = append(deck, GetActionCardDeck()...)

	// Rent Cards (55-67, 13)
	deck = append(deck, GetRentCardDeck()...)

	// Property Cards (68-95, 28)
	deck = append(deck, GetPropertyCardDeck()...)

	// Property Wildcards
	deck = append(deck, GetWildCardDeck()...)

	if len(deck) > 256 {
		// something is wrong with individual decks of cards types, cannot fit in uint8
		return nil
	}

	if len(deck) != 106 {
		// according to current logic of game, there should be 106 cards in total
		println("Error: deck size is", len(deck), "but expected 106")
		return nil
	}

	// changing ID assignment from sequence to something else will cause issues where we are finding property cards for arrangement
	// never assign 0 as card id, it can break system logic
	for i := range deck {
		deck[i].SetID(uint8(i + 1))
	}
	ShuffleDeck(deck)

	return deck
}

func ShuffleDeck(deck []Card) {
	n := len(deck)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
}

var colorSetRequirements = map[PropertyColor]int{
	Brown:     2,
	DarkBlue:  2,
	Green:     3,
	Red:       3,
	Yellow:    3,
	Orange:    3,
	Pink:      3,
	LightBlue: 3,
	Railroad:  4,
	Utility:   2,
}

func GetColorSetRequirement(color PropertyColor) int {
	if req, exists := colorSetRequirements[color]; exists {
		return req
	}
	return 0 // Default if color not found
}
