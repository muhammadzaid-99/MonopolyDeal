package game

import (
	c "cashdeal/models/cards"
	"slices"
)

func (g *Game) DoColorMatchRentColors(color c.PropertyColor, card *c.RentCard) bool {
	rentCardColors := card.GetColors()

	if rentCardColors[0] != c.Wild {
		return slices.Contains(rentCardColors, color)
	}
	return true
}

// func HandleRent(player *Player, target *Player, color c.PropertyColor, card c.RentCard, rentMultiplier uint8) bool {
// 	// TODO: need to handle rent for all players in game

// 	if DoColorMatchRentColors(color, card) {
// 		rent := player.GetRentOnColor(color)
// 		if rent != c.MoneyZero {
// 			return target.TransferMoney(player, player.GetRentOnColor(color)*rentMultiplier)

// 		}
// 	}
// 	return false
// }
