package game

// func (g *Game) HandleDealBreaker(player *Player, target *Player, color c.PropertyColor) bool {
// 	return target.TransferPropertySet(color, player)
// }

// func (g *Game) HandleDebtCollect(player *Player, target *Player, moneyCards []c.Card) bool {
// 	return target.TransferMoney(player, c.Money5M, moneyCards)
// }

// func (g *Game) HandleForcedDeal(player *Player, target *Player, playerCard c.Card, targetCard c.Card) bool {
// 	// TODO: here we have to allow target to rearrange

// 	if target.TransferPropertyIDBased(targetCard, player, false) {
// 		if player.TransferPropertyIDBased(playerCard, target, false) {
// 			return true
// 		} else {
// 			// Rollback the first transfer if the second fails
// 			player.TransferPropertyIDBased(targetCard, target, false)
// 			return false
// 		}
// 	}

// 	// TODO: another problem here is that it is possible for set complete and extra cards in unorganized pile which can be taken
// 	// if target.HasCompleteSet(targetColor) || player.HasCompleteSet(playerColor) {
// 	// 	return false
// 	// }

// 	// if target.HasProperty(targetColor) && player.HasProperty(playerColor) {
// 	// 	target.TransferProperty(targetColor, player)
// 	// 	player.TransferProperty(playerColor, target)
// 	// 	return true
// 	// }
// 	return false
// }

// func (g *Game) HandleSlyDeal(player *Player, target *Player, targetCard c.Card) bool {

// 	return target.TransferPropertyIDBased(targetCard, player, false)
// 	// if target.HasCompleteSet(color) {
// 	// 	return false
// 	// }
// 	// return target.TransferProperty(color, player)
// }

// func (g *Game) HandleBirthday(player *Player, targets []*Player) {
// 	for _, target := range targets {
// 		target.TransferMoney(player, c.Money2M)
// 	}
// }

func (g *Game) HandleDealBreaker(player *Player2, target *Player2, PileID string) bool {
	return target.TransferPropertySet(PileID, player)
}
