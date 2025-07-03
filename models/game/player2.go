package game

import (
	c "cashdeal/models/cards"
	"fmt"
	"slices"
	"sync"

	"github.com/google/uuid"
)

type PropertyPile struct {
	ID       string          `json:"id,omitempty"`
	Color    c.PropertyColor `json:"color,omitempty"`
	Cards    []c.Card        `json:"cards,omitempty"`
	Complete bool            `json:"complete,omitempty"`
	House    bool            `json:"house,omitempty"`
	Hotel    bool            `json:"hotel,omitempty"`
	Rent     uint8           `json:"rent,omitempty"`
}

type Player2 struct {
	ID            string
	Name          string
	HandCards     map[uint8]c.Card
	LooseCards    []c.Card                 // pile id empty string
	BankCards     []c.Card                 // Can contain MoneyCard or ActionCard
	PropertyPiles map[string]*PropertyPile // keys start from 1
	IsReady       bool
	ActionMessage map[string]string
	Mutex         sync.Mutex
}

func (p *Player2) NewPile() *PropertyPile {
	pile := &PropertyPile{
		ID: fmt.Sprintf("pile-%s", uuid.NewString()),
	}
	p.PropertyPiles[pile.ID] = pile
	return pile
}

func (p *Player2) DeletePile(PileID string) bool {
	if pile, ok := p.PropertyPiles[PileID]; ok {
		if len(pile.Cards) == 0 {
			delete(p.PropertyPiles, pile.ID)
		}
	}
	return false
}

func (p *Player2) ChangePileColor(PileID string, color c.PropertyColor) bool {
	if pile, ok := p.PropertyPiles[PileID]; ok {
		if len(pile.Cards) == 0 && color != c.Wild {
			pile.Color = color
			return true
		}
	}
	return false
}

func (p *Player2) DeleteEmptyPiles() {
	for _, pile := range p.PropertyPiles {
		if len(pile.Cards) == 0 {
			delete(p.PropertyPiles, pile.ID)
		}
	}
}

func (p *Player2) CanPay() bool {
	if len(p.BankCards) > 0 {
		return true
	}

	for _, c := range p.LooseCards {
		if c.GetMoneyValue() > 0 {
			return true
		}
	}

	for _, pile := range p.PropertyPiles {
		if pile.Rent > 0 {
			return true
		}
	}
	return false
}

func (p *Player2) AddToBank(card c.Card) bool {
	switch card.GetType() {
	case c.CardTypeMoney, c.CardTypeAction, c.CardTypeRent:
		p.BankCards = append(p.BankCards, card)
		return true
	default:
		return false
	}
}

func (p *Player2) HasInHand(cardID uint8) bool {
	_, exists := p.HandCards[cardID]
	return exists
}

func (p *Player2) AddToHand(card c.Card) {
	p.Mutex.Lock()
	p.HandCards[card.GetID()] = card
	p.Mutex.Unlock()
}

func (p *Player2) GetFromHand(cardID uint8) (c.Card, bool) {
	if p.HasInHand(cardID) {
		card := p.HandCards[cardID]
		delete(p.HandCards, cardID)
		return card, true
	}
	return nil, false
}

func (p *Player2) GetProperty(cardID uint8) (card c.Card, ok bool, pile *PropertyPile, commit func()) {
	// return func() is used to commit, bool to make sure only one time commit
	committed := false
	for i, card := range p.LooseCards {
		if card.GetID() == cardID {
			return card, true, nil, func() {
				if !committed {
					p.LooseCards = slices.Delete(p.LooseCards, i, i+1)
					committed = true
				}
			}
		}
	}

	for _, pile := range p.PropertyPiles {
		if len(pile.Cards) > 0 {
			for i, card := range pile.Cards {
				if card.GetID() == cardID {
					if pile.House {
						if ac, ok := card.(*c.ActionCard); ok {
							if ac.GetActionType() == c.ActionHouse && pile.Hotel {
								return nil, false, nil, func() {}
							}
						} else {
							return nil, false, nil, func() {}
						}
					}
					return card, true, pile, func() {
						if !committed {
							pile.Cards = slices.Delete(pile.Cards, i, i+1)
							p.DeletePile(pile.ID) // this will try to delete pile if its empty now
							p.UpdateRents(pile.ID)
							committed = true
						}
					}
				}
			}
		}
	}

	return nil, false, nil, func() {}
}

// wild case not handled, loose case
func (p *Player2) AddToProperty(PileID string, card c.Card) bool {
	defer p.UpdateRents(PileID)
	if PileID == "" {
		if slices.Contains([]c.CardType{c.CardTypeProperty, c.CardTypeWildcard}, card.GetType()) {
			p.LooseCards = append(p.LooseCards, card)
			return true
		} else if card.GetType() == c.CardTypeAction {
			if ac, ok := card.(*c.ActionCard); ok && slices.Contains([]c.ActionCardType{c.ActionHouse, c.ActionHotel}, ac.GetActionType()) {
				p.LooseCards = append(p.LooseCards, card)
				return true
			}
		}
	} else if pile, ok := p.PropertyPiles[PileID]; ok {
		switch card.GetType() {
		case c.CardTypeProperty:
			if pile.Complete {
				return false
			}
			fmt.Println("Going to add property card to ", PileID)
			if pc, ok := card.(*c.PropertyCard); ok && pc.GetColor() == pile.Color {
				fmt.Println("done", ok)
				pile.Cards = append(pile.Cards, card)
				return true
			}
		case c.CardTypeWildcard:
			if pile.Complete {
				return false
			}
			fmt.Println("Going to add wildcard to ", PileID)
			if wc, ok := card.(*c.Wildcard); ok && slices.Contains(wc.GetColors(), pile.Color) || slices.Contains(wc.GetColors(), c.Wild) {
				fmt.Println("done", ok)
				pile.Cards = append(pile.Cards, card)
				return true
			}

		case c.CardTypeAction:
			fmt.Println("Going to add action card to ", PileID)
			if ac, ok := card.(*c.ActionCard); ok && !IsARailroadOrUtility(pile.Color) {
				if ac.GetActionType() == c.ActionHouse && pile.Complete && !pile.House {
					fmt.Println("done", ok)
					pile.Cards = append(pile.Cards, card)
					return true
				}
				if ac.GetActionType() == c.ActionHotel && pile.House && !pile.Hotel {
					fmt.Println("done", ok)
					pile.Cards = append(pile.Cards, card)
					return true
				}
			}
		}
	}
	return false
}

func (p *Player2) GetRentOnPile(PileID string) (rent c.MoneyValue, completion uint8) {
	if PileID == "" {
		return c.MoneyZero, 0
	}
	if pile, ok := p.PropertyPiles[PileID]; ok {
		req := c.GetColorSetRequirement(pile.Color)
		count := len(pile.Cards)
		if count >= req+2 {
			return c.GetRentValuesForColor(pile.Color)[count-3] + HotelRent, 3
		} else if count >= req+1 {
			return c.GetRentValuesForColor(pile.Color)[count-2] + HouseRent, 2
		} else if count >= req {
			return c.GetRentValuesForColor(pile.Color)[count-1], 1
		} else if count > 0 {
			return c.GetRentValuesForColor(pile.Color)[count-1], 0
		}
	}
	return c.MoneyZero, 0
}

func (p *Player2) HasAnyCompleteSet() bool {
	for _, pile := range p.PropertyPiles {
		if pile.Complete {
			return true
		}
	}
	return false
}

func (p *Player2) HasAnyIncompleteSet() bool {
	if len(p.LooseCards) > 0 {
		return true
	}
	for _, pile := range p.PropertyPiles {
		if len(pile.Cards) > 0 && !pile.Complete {
			return true
		}
	}
	return false
}

func (p *Player2) HasCompleteSet(PileID string) bool {
	if pile, ok := p.PropertyPiles[PileID]; ok {
		return pile.Complete
	}
	return false
}

func (p *Player2) GetHandCount() uint8 {
	return uint8(len(p.HandCards))
}

func (p *Player2) IsHandEmpty() bool {
	return p.GetHandCount() == 0
}

func (p *Player2) HasProperty(color c.PropertyColor) bool {
	for _, pile := range p.PropertyPiles {
		if len(pile.Cards) > 0 && pile.Color == color {
			return true
		}
	}
	return false
}

func (p *Player2) GetCardForPayment(cardID uint8) (c.Card, bool) {
	index := slices.IndexFunc(p.BankCards, func(b c.Card) bool { return b.GetID() == cardID })
	if index != -1 {
		card := p.BankCards[index]
		p.BankCards = slices.Delete(p.BankCards, index, index+1)
		return card, true
	}
	if card, ok, _, commit := p.GetProperty(cardID); ok {
		if card.GetMoneyValue() > 0 {
			commit()
			return card, true
		}
	}
	return nil, false
}

func (p *Player2) UpdateRents(PileID string) {
	rent, completion := p.GetRentOnPile(PileID)
	if pile, ok := p.PropertyPiles[PileID]; ok {
		pile.Rent = uint8(rent)
		pile.Complete = (completion >= 1)
		pile.House = (completion >= 2)
		pile.Hotel = (completion >= 3)
	}
}

func (p *Player2) TransferPropertySet(PileID string, target *Player2) bool {
	if pile, ok := p.PropertyPiles[PileID]; ok {
		if pile.Complete {
			target.PropertyPiles[pile.ID] = pile
			delete(p.PropertyPiles, pile.ID)
			return true
		}
	}

	return false
}

func (p *Player2) HasRequiredPropertySets() bool {
	count := 0
	for _, pile := range p.PropertyPiles {
		if pile.Complete {
			count++
		}
	}
	return count >= RequiredPropertySets
}
