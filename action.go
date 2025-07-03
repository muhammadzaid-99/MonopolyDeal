package main

import (
	c "cashdeal/models/cards"
	"cashdeal/models/game"
	"fmt"
)

func (r *Room) HandleActionCardPlay(card c.Card) bool {

	switch card.GetType() {
	case c.CardTypeAction:
		if ac, ok := card.(*c.ActionCard); ok {
			switch ac.GetActionType() {
			case c.ActionBirthday:
				HasPaid := make(map[*PlayerConn]bool)
				for _, p := range r.Players {
					if p.ID != r.Game.CurrentTurn.ID {
						r.Game.PushNewMessage(fmt.Sprintf("%s is required to pay $%v to %s for their Birthday! ", p.Name, c.Money2M, r.Game.CurrentTurn.Name))

						r.SetPlayerActionMessage(p, map[string]string{
							"type":   "action",
							"action": "pay",
						})

						r.SetPlayerActionMessage(r.Players[r.Game.CurrentTurn.ID], map[string]string{
							"type":   "action",
							"action": "wait",
						})
						HasPaid[p] = false
					}
				}
				r.PendingAction = &PendingBirthday{
					BasePayable: c.Money2M,
					PaidMoney:   make(map[*PlayerConn]uint8),
					HasPaid:     HasPaid,
				}
				return true

			case c.ActionPassGo:
				r.SetPlayerActionMessage(r.Players[r.Game.CurrentTurn.ID], map[string]string{
					"type":   "action",
					"action": "draw-cards",
				})
				r.Game.PushNewMessage(fmt.Sprintf("%s will draw %v cards for 'Pass Go'", r.Game.CurrentTurn.Name, game.CardsDrawnPassGo))
				r.PendingAction = &PendingPassGo{}
				return true

			case c.ActionDebtCollect:
				r.SetPlayerActionMessage(r.Players[r.Game.CurrentTurn.ID], map[string]string{
					"type":   "action",
					"action": "select-an-opponent",
				})
				r.PendingAction = &PendingDebtCollector{
					BasePayable:  c.Money5M,
					Step:         "selection",
					TargetPlayer: nil,
				}
				return true

			case c.ActionHotel, c.ActionHouse:
				return false

			case c.ActionDealBreaker:
				for _, p := range r.Players {
					if p.HasAnyCompleteSet() && p.ID != r.Game.CurrentTurn.ID {
						r.SetPlayerActionMessage(r.Players[r.Game.CurrentTurn.ID], map[string]string{
							"type":   "action",
							"action": "select-opponent-property-complete-set",
						})
						r.PendingAction = &PendingDealBreaker{
							Step:         "selection",
							TargetPlayer: nil,
						}
						return true
					}
				}
				return false

			case c.ActionDoubleRent:
				if a, ok := r.PendingAction.(*PendingRent); ok {
					a.RentMultiplier *= 2
					return true
				}
				return false

			case c.ActionForcedDeal:
				if !r.Game.CurrentTurn.HasAnyIncompleteSet() {
					return false
				}
				fallthrough
			case c.ActionSlyDeal:
				for _, p := range r.Players {
					if p.HasAnyIncompleteSet() && p.ID != r.Game.CurrentTurn.ID {
						switch ac.GetActionType() {
						case c.ActionForcedDeal:
							r.PendingAction = &PendingForcedDeal{
								Step:         "selection",
								TargetPlayer: nil,
							}
						case c.ActionSlyDeal:
							r.PendingAction = &PendingSlyDeal{
								Step:         "selection",
								TargetPlayer: nil,
							}
						}
						r.SetPlayerActionMessage(r.Players[r.Game.CurrentTurn.ID], map[string]string{
							"type":   "action",
							"action": "select-opponent-property",
						})
						return true
					}
				}
				return false

			case c.ActionJustSayNo:
				// this will work if some pending action calls it. other cases should make replace pending action
				tc := r.Game.TopCard
				if tc.GetType() == c.CardTypeAction {
					if jsn, ok := tc.(*c.ActionCard); ok {
						if jsn.GetActionType() == c.ActionJustSayNo && r.PendingAction != nil {
							// do not replace pending action, so it will continue
							return true
						}
					}
				}
				return false
			}
		}
	case c.CardTypeRent:
		if rc, ok := card.(*c.RentCard); ok {
			for _, pile := range r.Game.CurrentTurn.PropertyPiles {
				fmt.Println("Checking for (color, rent, rent-card-match?):", pile.Color, pile.Cards, r.Game.DoColorMatchRentColors(pile.Color, rc))
				if pile.Rent > 0 && r.Game.DoColorMatchRentColors(pile.Color, rc) {
					r.PendingAction = &PendingRent{
						BasePayable:    c.MoneyZero,
						RentMultiplier: 1,
						PaidMoney:      make(map[*PlayerConn]uint8),
						HasPaid:        make(map[*PlayerConn]bool),
						TargetPlayer:   nil,
						RentCard:       rc,
						Step:           "selection",
						IsForOnePlayer: len(rc.GetColors()) != 2,
					}
					fmt.Println("colors in rent card are ", rc.GetColors(), len(rc.GetColors()))
					if len(rc.GetColors()) != 2 {
						r.SetPlayerActionMessage(r.Players[r.Game.CurrentTurn.ID], map[string]string{
							"type":   "action",
							"action": "select-your-property-set-and-opponent",
						})
					} else {
						r.SetPlayerActionMessage(r.Players[r.Game.CurrentTurn.ID], map[string]string{
							"type":   "action",
							"action": "select-your-property-complete-set",
						})
					}
					return true
				}
			}
		}
		return false
	}

	return false
}
