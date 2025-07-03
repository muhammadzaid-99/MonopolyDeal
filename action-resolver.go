package main

import (
	c "cashdeal/models/cards"
	"cashdeal/models/game"
	"fmt"
)

type PendingAction interface {
	Resolve(r *Room, p *PlayerConn, msg WSMessage)
}

type PendingDealBreaker struct {
	TargetPile   *game.PropertyPile
	TargetPlayer *PlayerConn
	Step         string
}

func (pd *PendingDealBreaker) Resolve(room *Room, p *PlayerConn, msg WSMessage) {
	switch pd.Step {
	case "selection":
		if p.ID != room.Game.CurrentTurn.ID {
			return
		}
		pile_id := msg.PropertyPileID
		targetPlayerID := msg.TargetPlayerID
		if pile_id == "" || targetPlayerID == "" {
			return
		}
		if targetPlayerID == p.ID {
			// cannot target self
			return
		}
		if player, ok := room.Players[targetPlayerID]; ok {
			if pile, ok := player.PropertyPiles[pile_id]; ok && pile.Complete {
				pd.TargetPile = pile
				pd.TargetPlayer = player
				room.Game.PushNewMessage(fmt.Sprintf("%s requested a %s property set from %s", p.Name, pile.Color, player.Name))
				room.SetPlayerActionMessage(p, map[string]string{
					"type":   "action",
					"action": "wait",
				})
				room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
					"type":   "action",
					"action": "continue",
				})
				pd.Step = "reaction"
			}
		}

	case "reaction":
		if p.ID != pd.TargetPlayer.ID {
			return
		}
		if card, ok := pd.TargetPlayer.GetFromHand(msg.CardID); ok {
			if ac, ok := card.(*c.ActionCard); ok {
				if ac.GetActionType() == c.ActionJustSayNo {
					room.Game.MoveToDiscardPile(ac)
					room.Game.PushNewMessage(fmt.Sprintf("%s played 'Just Say No' to %s's 'Deal Breaker'", p.Name, room.Game.CurrentTurn.Name))
					room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
						"type":   "action",
						"action": "continue",
					})
					room.SetPlayerActionMessage(p, map[string]string{
						"type":   "action",
						"action": "wait",
					})
					pd.Step = "cancelled"
					return
				}
			}
			pd.TargetPlayer.AddToHand(card)
		} else {
			// no response with Just Say No, continue deal breaking
			if msg.Type == "continue" {
				room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
					"type":   "action",
					"action": "clear",
				})
				room.SetPlayerActionMessage(p, map[string]string{
					"type":   "action",
					"action": "clear",
				})
				room.Game.HandleDealBreaker(room.Game.CurrentTurn, &pd.TargetPlayer.Player2, pd.TargetPile.ID)
				room.Game.PushNewMessage(fmt.Sprintf("Deal Breaker: %s stole a %s property set from %s", room.Game.CurrentTurn.Name, pd.TargetPile.Color, p.Name))
				room.Game.PushNewMessage("'Deal Breaker' was completed!")
				room.PendingAction = nil
				room.CheckWin(room.Game.CurrentTurn)
			}
		}
	case "cancelled":
		if p.ID != room.Game.CurrentTurn.ID {
			return
		}
		if room.PlayCard(p.ID, msg.CardID, DiscardPile) {
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "continue",
			})
			room.SetPlayerActionMessage(p, map[string]string{
				"type":   "action",
				"action": "wait",
			})
			room.Game.PushNewMessage(fmt.Sprintf("%s played 'Just Say No' to %s's 'Just Say No'", p.Name, pd.TargetPlayer.Name))
			pd.Step = "reaction"
			return
		}
		if msg.Type == "continue" {
			room.PendingAction = nil
			room.SetPlayerActionMessage(p, map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.Game.PushNewMessage("'Deal Breaker' was unsuccessful!")
		}
	}
}

type PendingSlyDeal struct {
	TargetPlayer *PlayerConn
	TargetPile   *game.PropertyPile
	TargetCard   c.Card
	TargetCommit func()
	CardID       uint8
	Step         string
}

func (pd *PendingSlyDeal) Resolve(room *Room, p *PlayerConn, msg WSMessage) {
	switch pd.Step {
	case "selection":
		if p.ID != room.Game.CurrentTurn.ID {
			return
		}
		cardID := msg.CardID
		targetPlayerID := msg.TargetPlayerID
		if targetPlayerID == "" {
			return
		}
		if targetPlayerID == p.ID {
			// cannot target self
			return
		}
		if player, ok := room.Players[targetPlayerID]; ok {
			pd.TargetPlayer = player
			if card, ok, pile, commit := pd.TargetPlayer.GetProperty(cardID); ok {
				if pile != nil && pile.Complete {
					return
				}
				pd.TargetPile = pile
				pd.TargetCard = card
				pd.TargetCommit = commit
				// p.AddToProperty(c.Wild, card)
				// commit()
				room.Game.PushNewMessage(fmt.Sprintf("%s requested %s property from %s", p.Name, pd.TargetCard.GetName(), player.Name))
				room.SetPlayerActionMessage(p, map[string]string{
					"type":   "action",
					"action": "wait",
				})
				room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
					"type":   "action",
					"action": "continue",
				})
				pd.Step = "reaction"
			}

		}
	case "reaction":
		if p.ID != pd.TargetPlayer.ID {
			return
		}
		if card, ok := pd.TargetPlayer.GetFromHand(msg.CardID); ok {
			if ac, ok := card.(*c.ActionCard); ok {
				if ac.GetActionType() == c.ActionJustSayNo {
					room.Game.MoveToDiscardPile(ac)
					room.Game.PushNewMessage(fmt.Sprintf("%s played 'Just Say No' to %s's 'Sly Deal'", p.Name, room.Game.CurrentTurn.Name))
					room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
						"type":   "action",
						"action": "continue",
					})
					room.SetPlayerActionMessage(p, map[string]string{
						"type":   "action",
						"action": "wait",
					})
					pd.Step = "cancelled"
					return
				}
			}
			pd.TargetPlayer.AddToHand(card)
		} else {
			// no response with Just Say No, continue with sly dealing
			if msg.Type == "continue" {
				pd.TargetCommit()
				room.Game.CurrentTurn.AddToProperty("", pd.TargetCard)
				room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
					"type":   "action",
					"action": "clear",
				})
				room.SetPlayerActionMessage(p, map[string]string{
					"type":   "action",
					"action": "clear",
				})
				room.Game.PushNewMessage(fmt.Sprintf("Sly Deal: %s stole %s property from %s", room.Game.CurrentTurn.Name, pd.TargetCard.GetName(), p.Name))
				room.Game.PushNewMessage("'Sly Deal' was completed!")
				room.PendingAction = nil
			}
		}
	case "cancelled":
		if p.ID != room.Game.CurrentTurn.ID {
			return
		}
		if room.PlayCard(p.ID, msg.CardID, DiscardPile) {
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "continue",
			})
			room.SetPlayerActionMessage(p, map[string]string{
				"type":   "action",
				"action": "wait",
			})
			room.Game.PushNewMessage(fmt.Sprintf("%s played 'Just Say No' to %s's 'Just Say No'", p.Name, pd.TargetPlayer.Name))
			pd.Step = "reaction"
			return
		}
		if msg.Type == "continue" {
			room.PendingAction = nil
			room.SetPlayerActionMessage(p, map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.Game.PushNewMessage("'Sly Deal' was unsuccessful!")
		}
	}
}

type PendingForcedDeal struct {
	PlayerCardID uint8
	PlayerPile   *game.PropertyPile
	PlayerCard   c.Card
	PlayerCommit func()
	TargetPlayer *PlayerConn
	TargetCardID uint8
	TargetPile   *game.PropertyPile
	TargetCard   c.Card
	TargetCommit func()
	Step         string
}

func (pd *PendingForcedDeal) Resolve(room *Room, p *PlayerConn, msg WSMessage) {
	switch pd.Step {
	case "selection":
		if p.ID != room.Game.CurrentTurn.ID {
			return
		}
		pd.TargetCardID = msg.CardID
		targetPlayerID := msg.TargetPlayerID
		fmt.Println("Got player ID and card ID", targetPlayerID, msg.CardID)
		if targetPlayerID == "" {
			return
		}
		if targetPlayerID == p.ID {
			// cannot target self
			fmt.Println("Targetting self")
			return
		}
		if !p.HasAnyIncompleteSet() {
			fmt.Println("Does not have any incomplete set")
			return
		}
		if player, ok := room.Players[targetPlayerID]; ok {
			pd.TargetPlayer = player
			fmt.Println("Target player set")
			if card, ok, pile, commit := pd.TargetPlayer.GetProperty(pd.TargetCardID); ok {
				if pile != nil && pile.Complete {
					return
				}
				fmt.Println("Opponent Property acquired")
				pd.TargetPile = pile
				pd.TargetCard = card
				pd.TargetCommit = commit
				// p.AddToProperty(c.Wild, card)
				// commit()
				room.SetPlayerActionMessage(p, map[string]string{
					"type":   "action",
					"action": "select-your-property",
				})
				fmt.Println("Now player has to select their property")
				pd.Step = "provision"
			}

		}
	case "provision":
		if p.ID != room.Game.CurrentTurn.ID {
			fmt.Println("Only player is allowed")
			return
		}
		pd.PlayerCardID = msg.CardID
		fmt.Println("Player card ID set to", msg.CardID)
		if card, ok, pile, commit := p.GetProperty(pd.PlayerCardID); ok {
			if pile != nil && pile.Complete {
				return
			}
			fmt.Println("Player property acquired")
			pd.PlayerPile = pile
			pd.PlayerCard = card
			pd.PlayerCommit = commit
			room.SetPlayerActionMessage(p, map[string]string{
				"type":   "action",
				"action": "wait",
			})
			fmt.Println("Player free now")
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "continue",
			})
			room.Game.PushNewMessage(fmt.Sprintf("%s requested %s property from %s in exchange of their %s property", p.Name, pd.TargetCard.GetName(), pd.TargetPlayer.Name, pd.PlayerCard.GetName()))
			fmt.Println("Opponent has to continue")
			// pd.TargetPlayer.AddToProperty(c.Wild, card)
			// commit()
			pd.Step = "reaction"
		}
	case "reaction":
		if p.ID != pd.TargetPlayer.ID {
			fmt.Println("Opponent can respond only")
			return
		}
		if card, ok := pd.TargetPlayer.GetFromHand(msg.CardID); ok {
			if ac, ok := card.(*c.ActionCard); ok {
				if ac.GetActionType() == c.ActionJustSayNo {
					room.Game.MoveToDiscardPile(ac)
					fmt.Println("Opponent played just say no and is cleared, cancelled")
					room.Game.PushNewMessage(fmt.Sprintf("%s played 'Just Say No' to %s's 'Forced Deal'", p.Name, room.Game.CurrentTurn.Name))
					room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
						"type":   "action",
						"action": "continue",
					})
					room.SetPlayerActionMessage(p, map[string]string{
						"type":   "action",
						"action": "wait",
					})
					pd.Step = "cancelled"
					return
				}
			}
			pd.TargetPlayer.AddToHand(card)
			fmt.Println("Opponent played invalid card")
		} else {
			// no response with Just Say No, continue with force dealing
			// ...
			// ...
			// 	HAVE TO TAKE CARE OF RACE CONDITIONS
			// ...
			// ..
			if msg.Type == "continue" {
				fmt.Println("Forced dealing")
				pd.PlayerCommit()
				pd.TargetCommit()
				pd.TargetPlayer.AddToProperty("", pd.PlayerCard)
				room.Game.CurrentTurn.AddToProperty("", pd.TargetCard)
				fmt.Println("Done")
				room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
					"type":   "action",
					"action": "wait",
				})
				room.SetPlayerActionMessage(p, map[string]string{
					"type":   "action",
					"action": "arrange",
				})
				room.Game.PushNewMessage(fmt.Sprintf("Forced Deal: %s stole %s property from %s in exchange of their %s property", room.Game.CurrentTurn.Name, pd.TargetCard.GetName(), p.Name, pd.PlayerCard.GetName()))
				room.Game.PushNewMessage(fmt.Sprintf("%s is arranging their transferred property", p.Name))
				pd.Step = "arrange"
			}
		}
	case "cancelled":
		if p.ID != room.Game.CurrentTurn.ID {
			return
		}
		if room.PlayCard(p.ID, msg.CardID, DiscardPile) {
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "continue",
			})
			room.SetPlayerActionMessage(p, map[string]string{
				"type":   "action",
				"action": "wait",
			})
			room.Game.PushNewMessage(fmt.Sprintf("%s played 'Just Say No' to %s's 'Just Say No'", p.Name, pd.TargetPlayer.Name))
			pd.Step = "reaction"
			fmt.Println("Player played just say no")
			return
		}
		if msg.Type == "continue" {
			// pd.Step = "arrange"
			room.SetPlayerActionMessage(p, map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.Game.PushNewMessage("'Forced Deal' was unsuccessful!")
			room.PendingAction = nil
		}
	case "arrange":
		if p.ID != pd.TargetPlayer.ID {
			return
		}
		if msg.Type == "create-pile" {
			if len(pd.TargetPlayer.PropertyPiles) < game.MaxPilesAllowed {
				pd.TargetPlayer.NewPile()
				room.BroadcastAllGameState()
			}
		}
		if msg.Type == "change-pile-color" {
			pd.TargetPlayer.ChangePileColor(msg.PropertyPileID, msg.PropertyColor)
			room.BroadcastAllGameState()
		}
		if msg.Type == "arrange-property" && msg.CardID == pd.PlayerCardID {
			if card, ok, _, commit := pd.TargetPlayer.GetProperty(msg.CardID); ok {
				if pd.TargetPlayer.AddToProperty(msg.PropertyPileID, card) {
					commit()
					room.Game.PushNewMessage(fmt.Sprintf("%s moved their %s card", pd.TargetPlayer.Name, card.GetName()))
				}
			}
		}
		if msg.Type == "continue" {
			room.PendingAction = nil
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.Game.PushNewMessage("'Forced Deal' was completed!")
		}
	}
}

type PendingBirthday struct {
	BasePayable c.MoneyValue
	PaidMoney   map[*PlayerConn]uint8
	HasPaid     map[*PlayerConn]bool
}

func (pd *PendingBirthday) Resolve(room *Room, p *PlayerConn, msg WSMessage) {

	if p.ID == room.Game.CurrentTurn.ID {
		return
	}

	if pd.HasPaid[p] {
		return
	}

	if !p.CanPay() {
		if msg.Type == "continue" {
			pd.HasPaid[p] = true
			room.Game.PushNewMessage(fmt.Sprintf("%s paid a total of $%v to %s. They cannot pay anymore.", p.Name, pd.PaidMoney[p], room.Game.CurrentTurn.Name))
			room.SetPlayerActionMessage(p, map[string]string{
				"type":   "action",
				"action": "clear",
			})
		}
	} else {
		if card, ok := p.GetCardForPayment(msg.CardID); ok {
			switch card.GetType() {
			case c.CardTypeMoney, c.CardTypeAction, c.CardTypeRent:
				room.Game.CurrentTurn.AddToBank(card)
			case c.CardTypeProperty, c.CardTypeWildcard:
				room.Game.CurrentTurn.AddToProperty("", card)
			}
			pd.PaidMoney[p] += uint8(card.GetMoneyValue())
			room.Game.PushNewMessage(fmt.Sprintf("%s paid $%v to %s with %s", p.Name, card.GetMoneyValue(), room.Game.CurrentTurn.Name, card.GetName()))
			if pd.PaidMoney[p] >= uint8(pd.BasePayable) {
				pd.HasPaid[p] = true
				room.SetPlayerActionMessage(p, map[string]string{
					"type":   "action",
					"action": "clear",
				})
				room.Game.PushNewMessage(fmt.Sprintf("%s paid a total of $%v to %s", p.Name, pd.PaidMoney[p], room.Game.CurrentTurn.Name))
			}
		} else if card, ok := p.GetFromHand(msg.CardID); ok && card.GetType() == c.CardTypeAction {
			if jsn, ok := card.(*c.ActionCard); ok && jsn.GetActionType() == c.ActionJustSayNo && pd.PaidMoney[p] == 0 {
				room.Game.MoveToDiscardPile(card)
				pd.HasPaid[p] = true
				room.Game.PushNewMessage(fmt.Sprintf("%s played 'Just Say No' to %s", p.Name, room.Game.CurrentTurn.Name))
				room.SetPlayerActionMessage(p, map[string]string{
					"type":   "action",
					"action": "clear",
				})
			}
		}
	}

	for _, paid := range pd.HasPaid {
		if !paid {
			return
		}
	}

	// everyone has paid
	room.PendingAction = nil
	room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
		"type":   "action",
		"action": "clear",
	})
	room.Game.PushNewMessage("'Its My Birthday' was completed!")
}

type PendingDebtCollector struct {
	BasePayable  c.MoneyValue
	PaidMoney    uint8
	TargetPlayer *PlayerConn
	Step         string
}

func (pd *PendingDebtCollector) Resolve(room *Room, p *PlayerConn, msg WSMessage) {
	switch pd.Step {
	case "selection":
		if p.ID != room.Game.CurrentTurn.ID {
			return
		}

		if pd.TargetPlayer == nil {
			targetPlayerID := msg.TargetPlayerID
			if targetPlayerID == p.ID {
				// cannot target self
				return
			}
			if player, ok := room.Players[targetPlayerID]; ok {
				pd.TargetPlayer = player
				pd.Step = "reaction"
				room.SetPlayerActionMessage(p, map[string]string{
					"type":   "action",
					"action": "wait",
				})
				room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
					"type":   "action",
					"action": "pay",
				})
				room.Game.PushNewMessage(fmt.Sprintf("%s requested $%v from %s", p.Name, pd.BasePayable, pd.TargetPlayer.Name))
			}
		}

	case "reaction":
		if p.ID == room.Game.CurrentTurn.ID {
			return
		}

		if !p.CanPay() {
			if msg.Type == "continue" {
				room.PendingAction = nil
				room.Game.PushNewMessage(fmt.Sprintf("%s paid a total of $%v to %s. They cannot pay anymore.", p.Name, pd.PaidMoney, room.Game.CurrentTurn.Name))
				room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
					"type":   "action",
					"action": "clear",
				})
				room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
					"type":   "action",
					"action": "clear",
				})
			}
		} else {
			if card, ok := p.GetCardForPayment(msg.CardID); ok {
				switch card.GetType() {
				case c.CardTypeMoney, c.CardTypeAction, c.CardTypeRent:
					room.Game.CurrentTurn.AddToBank(card)
				case c.CardTypeProperty, c.CardTypeWildcard:
					room.Game.CurrentTurn.AddToProperty("", card)
				}
				pd.PaidMoney += uint8(card.GetMoneyValue())
				room.Game.PushNewMessage(fmt.Sprintf("%s paid $%v to %s with %s", p.Name, card.GetMoneyValue(), room.Game.CurrentTurn.Name, card.GetName()))
				if pd.PaidMoney >= uint8(pd.BasePayable) {
					room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
						"type":   "action",
						"action": "clear",
					})
					room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
						"type":   "action",
						"action": "clear",
					})
					room.Game.PushNewMessage(fmt.Sprintf("%s paid a total of $%v to %s", p.Name, pd.PaidMoney, room.Game.CurrentTurn.Name))
					room.Game.PushNewMessage("'Debt Collector' was completed!")
					room.PendingAction = nil
				}
			} else if card, ok := p.GetFromHand(msg.CardID); ok && card.GetType() == c.CardTypeAction {
				if jsn, ok := card.(*c.ActionCard); ok && jsn.GetActionType() == c.ActionJustSayNo && pd.PaidMoney == 0 {
					room.Game.MoveToDiscardPile(card)
					room.Game.PushNewMessage(fmt.Sprintf("%s played 'Just Say No' to %s", p.Name, room.Game.CurrentTurn.Name))
					room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
						"type":   "action",
						"action": "wait",
					})
					room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
						"type":   "action",
						"action": "continue",
					})
					pd.Step = "cancelled"
				}
			}
		}
	case "cancelled":
		if p.ID != room.Game.CurrentTurn.ID {
			return
		}
		if room.PlayCard(p.ID, msg.CardID, DiscardPile) {
			pd.Step = "reaction"
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "pay",
			})
			room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
				"type":   "action",
				"action": "wait",
			})
			room.Game.PushNewMessage(fmt.Sprintf("%s played 'Just Say No' to %s's 'Just Say No'", p.Name, pd.TargetPlayer.Name))
			return
		}
		if msg.Type == "continue" {
			room.PendingAction = nil
			room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.Game.PushNewMessage("'Debt Collector' was unsuccessful!")
		}
	}

}

type PendingRent struct {
	BasePayable    c.MoneyValue
	PaidMoney      map[*PlayerConn]uint8
	HasPaid        map[*PlayerConn]bool
	IsForOnePlayer bool
	TargetPlayer   *PlayerConn
	RentCard       *c.RentCard
	PlayerPile     *game.PropertyPile
	Step           string
	RentMultiplier uint8
}

func (pd *PendingRent) Resolve(room *Room, p *PlayerConn, msg WSMessage) {
	switch pd.Step {
	case "selection":
		if p.ID != room.Game.CurrentTurn.ID {
			return
		}

		// try to play card, maybe its double the rent, it will change multiplier
		if pd.RentCard != nil {
			room.PlayCard(p.ID, msg.CardID, DiscardPile)
		}

		// if msg.PropertyPileID == "" || (pd.IsForOnePlayer && msg.TargetPlayerID == "") {
		// 	// give both at once
		// 	return
		// }

		pile_id := msg.PropertyPileID
		if pile_id != "" {
			if pile, ok := p.PropertyPiles[pile_id]; ok {
				if pd.RentCard != nil && room.Game.DoColorMatchRentColors(pile.Color, pd.RentCard) {
					pd.BasePayable, _ = p.GetRentOnPile(pile_id)
					pd.PlayerPile = pile
					fmt.Println("Base payable set to:", pd.BasePayable)
				}
			}
		}

		if pd.IsForOnePlayer {
			if pd.TargetPlayer == nil {
				targetPlayerID := msg.TargetPlayerID
				if targetPlayerID == p.ID {
					// cannot target self
					return
				}
				if player, ok := room.Players[targetPlayerID]; ok {
					pd.TargetPlayer = player
				}
			}
		}

		if pd.BasePayable != c.MoneyZero && (!pd.IsForOnePlayer || pd.TargetPlayer != nil) {
			room.SetPlayerActionMessage(p, map[string]string{
				"type":   "action",
				"action": "continue",
			})
			if pd.IsForOnePlayer {
				room.Game.PushNewMessage(fmt.Sprintf("%s will be expecting $%v rent on %s property from %s", p.Name, uint8(pd.BasePayable)*pd.RentMultiplier, pd.PlayerPile.Color, pd.TargetPlayer.Name))
			} else {
				room.Game.PushNewMessage(fmt.Sprintf("%s will be expecting $%v rent on %s property from everyone", p.Name, uint8(pd.BasePayable)*pd.RentMultiplier, pd.PlayerPile.Color))
			}
			fmt.Println("Asking player to continue or double the rent.")
			// here continue is sent to make sure if they want to continue or not
			pd.Step = "double"
		}

	case "double":
		if p.ID != room.Game.CurrentTurn.ID {
			return
		}
		// try to play card, maybe its double the rent, it will change multiplier
		if pd.RentCard != nil && room.PlayCard(p.ID, msg.CardID, DiscardPile) {
			room.Game.PushNewMessage(fmt.Sprintf("%s doubled the rent to %v", pd.TargetPlayer.Name, uint8(pd.BasePayable)*pd.RentMultiplier))
			fmt.Println("Rent doubled.")
			return
		} else if msg.Type == "continue" {
			fmt.Println("Rent not doubled.")
			if msg.Type != "continue" {
				fmt.Println("Continue not received.")
				return
			}
			pd.Step = "reaction"
			room.SetPlayerActionMessage(p, map[string]string{
				"type":   "action",
				"action": "wait",
			})
			fmt.Println("Player continued and is clear now.")
			if pd.IsForOnePlayer {
				room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
					"type":   "action",
					"action": "pay",
				})
				room.Game.PushNewMessage(fmt.Sprintf("%s is required to pay $%v rent to %s", pd.TargetPlayer.Name, uint8(pd.BasePayable)*pd.RentMultiplier, p.Name))
				fmt.Println("Single opponent has to pay now.")
				pd.HasPaid[pd.TargetPlayer] = false
			} else {
				for _, plr := range room.Players {
					if plr.ID != p.ID {
						room.SetPlayerActionMessage(plr, map[string]string{
							"type":   "action",
							"action": "pay",
						})
						room.Game.PushNewMessage(fmt.Sprintf("%s is required to pay $%v rent to %s", plr.Name, uint8(pd.BasePayable)*pd.RentMultiplier, p.Name))
						pd.HasPaid[plr] = false
					}
				}
				fmt.Println("All opponents have to pay now.")
			}
		}
	case "reaction":
		if p.ID == room.Game.CurrentTurn.ID {
			fmt.Println("Player tried to react.")
			return
		}

		if pd.IsForOnePlayer && p.ID != pd.TargetPlayer.ID {
			fmt.Println("Opponent (non-target) tried to react.")
			return
		}

		if pd.HasPaid[p] {
			fmt.Println("This opponent has already paid.")
			return
		}

		if !p.CanPay() {
			if msg.Type == "continue" {
				pd.HasPaid[p] = true
				fmt.Println("This opponent cannot pay. So has paid.")
				room.SetPlayerActionMessage(p, map[string]string{
					"type":   "action",
					"action": "clear",
				})
				room.Game.PushNewMessage(fmt.Sprintf("%s paid a total of $%v to %s. They cannot pay anymore.", p.Name, pd.PaidMoney[p], room.Game.CurrentTurn.Name))
			}
		} else {
			if card, ok := p.GetCardForPayment(msg.CardID); ok {
				switch card.GetType() {
				case c.CardTypeMoney, c.CardTypeAction, c.CardTypeRent:
					room.Game.CurrentTurn.AddToBank(card)
				case c.CardTypeProperty, c.CardTypeWildcard:
					room.Game.CurrentTurn.AddToProperty("", card)
				}
				fmt.Println("The opponent paid some with a card.")
				pd.PaidMoney[p] += uint8(card.GetMoneyValue())
				room.Game.PushNewMessage(fmt.Sprintf("%s paid $%v to %s with %s", p.Name, card.GetMoneyValue(), room.Game.CurrentTurn.Name, card.GetName()))
				if pd.PaidMoney[p] >= uint8(pd.BasePayable)*pd.RentMultiplier {
					pd.HasPaid[p] = true
					fmt.Println("The opponent has paid all.")

					room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
						"type":   "action",
						"action": "clear",
					})
					room.SetPlayerActionMessage(p, map[string]string{
						"type":   "action",
						"action": "clear",
					})
					room.Game.PushNewMessage(fmt.Sprintf("%s paid a total of $%v to %s", p.Name, pd.PaidMoney[p], room.Game.CurrentTurn.Name))
				}
			} else if card, ok := p.GetFromHand(msg.CardID); ok && card.GetType() == c.CardTypeAction {
				if jsn, ok := card.(*c.ActionCard); ok && jsn.GetActionType() == c.ActionJustSayNo && pd.PaidMoney[p] == 0 {
					room.Game.MoveToDiscardPile(card)
					if pd.IsForOnePlayer {
						pd.Step = "cancelled"
						room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
							"type":   "action",
							"action": "clear",
						})
						room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
							"type":   "action",
							"action": "continue",
						})
						fmt.Println("The opponent has paid all playing Just Say No. Cancelled.")
					} else {
						pd.HasPaid[p] = true
						fmt.Println("This one opponent has paid playing Just Say No.")
					}
					room.Game.PushNewMessage(fmt.Sprintf("%s played 'Just Say No' to %s", p.Name, room.Game.CurrentTurn.Name))
				}
			}
		}

		if pd.IsForOnePlayer && pd.HasPaid[pd.TargetPlayer] {
			fmt.Println("The opponet has paid all.")
			// everyone has paid
			room.Game.PushNewMessage("'RENT' was completed!")
			room.PendingAction = nil
			return
		}

		for _, paid := range pd.HasPaid {
			if !paid {
				fmt.Println("Someone has not paid yet.")
				return
			}
		}

		// everyone has paid
		fmt.Println("Everyone has paid.")
		room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
			"type":   "action",
			"action": "clear",
		})
		room.PendingAction = nil
		room.Game.PushNewMessage("'RENT' was completed!")
	case "cancelled":
		if p.ID != room.Game.CurrentTurn.ID {
			return
		}
		if room.PlayCard(p.ID, msg.CardID, DiscardPile) {
			pd.Step = "reaction"
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "pay",
			})
			room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
				"type":   "action",
				"action": "wait",
			})
			room.Game.PushNewMessage(fmt.Sprintf("%s played 'Just Say No' to %s's 'Just Say No'", p.Name, pd.TargetPlayer.Name))
			return
		}
		if msg.Type == "continue" {
			room.PendingAction = nil
			room.SetPlayerActionMessage(room.Players[room.Game.CurrentTurn.ID], map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.SetPlayerActionMessage(pd.TargetPlayer, map[string]string{
				"type":   "action",
				"action": "clear",
			})
			room.Game.PushNewMessage("'RENT' was unsuccessful!")
		}
	}

}

type PendingPassGo struct{}

func (pd *PendingPassGo) Resolve(room *Room, p *PlayerConn, msg WSMessage) {
	if p.ID != room.Game.CurrentTurn.ID {
		return
	}

	if msg.Type == "draw-cards" {
		for range game.CardsDrawnPassGo {
			card := room.Game.DrawCard()
			p.AddToHand(card)
		}
		room.Game.PushNewMessage(fmt.Sprintf("%s has drawn %v cards", p.Name, game.CardsDrawnPassGo))
		room.SetPlayerActionMessage(p, map[string]string{
			"type":   "action",
			"action": "clear",
		})
		room.PendingAction = nil
	}
}
