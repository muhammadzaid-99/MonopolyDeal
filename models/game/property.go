package game

import c "cashdeal/models/cards"

// func DoColorMatchWildColors(color c.PropertyColor, wc *c.Wildcard) bool {
// 	colors := wc.GetColors()
// 	if colors[0] != c.Wild {
// 		return slices.Contains(colors, color)
// 	}
// 	return true
// }

func IsARailroadOrUtility(color c.PropertyColor) bool {
	return color == c.Railroad || color == c.Utility
}

// func HandleAddToProperty(player *Player, color c.PropertyColor, card c.Card) bool {

// 	if player.HasCompleteSet(color) {
// 		return false
// 	}

// 	switch card.GetType() {
// 	case c.CardTypeProperty:
// 		if pc, ok := card.(*c.PropertyCard); !ok {
// 			return false
// 		} else if color != pc.GetColor() {
// 			return false
// 		}

// 	case c.CardTypeWildcard:
// 		if wc, ok := card.(*c.Wildcard); !ok {
// 			return false
// 		} else if !DoColorMatchWildColors(color, wc) {
// 			return false
// 		}
// 	}

// 	player.AddToProperty(color, card)
// 	return true
// }

// TODO -------------
// If a player has a house/hotel on the table that is not part of a full completed set,
//
//	then a Force Deal or Sly Deal can be used to steel it.
//
// See house/hotel question below for how a house/hotel can
//
//	be placed on the table by its self.
// func HandleAddHouseHotelToProperty(player *Player, color c.PropertyColor, card c.Card) bool {
// 	if IsARailroadOrUtility(color) || !player.HasCompleteSet(color) {
// 		return false
// 	}
// 	if pc, ok := card.(*c.ActionCard); !ok {
// 		return false
// 	} else {
// 		actype := pc.GetActionType()
// 		if actype == c.ActionHouse && !player.HasHouse(color) {
// 			return player.AddToProperty(color, card)
// 		} else if actype == c.ActionHotel && !player.HasHotel(color) {
// 			return player.AddToProperty(color, card)
// 		}
// 	}

// 	return false
// }
