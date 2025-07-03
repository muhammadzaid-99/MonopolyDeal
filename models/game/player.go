package game

// type Player struct {
// 	ID                 string
// 	Name               string
// 	HandCards          map[uint8]c.Card
// 	PropertyCards      map[c.PropertyColor][]c.Card // PropertyColor maps to a slice of PropertyCard or PropertyWildcard
// 	BankCards          []c.Card                     // Can contain MoneyCard or ActionCard
// 	PropertyRents      map[c.PropertyColor]c.MoneyValue
// 	PropertyCompletion map[c.PropertyColor]uint8
// 	IsReady            bool
// 	ActionMessage      map[string]string
// 	Mutex              sync.Mutex
// }

// func (p *Player) CanPay() bool {
// 	if len(p.BankCards) > 0 {
// 		return true
// 	}

// 	for _, clr := range c.GetAllColors() {
// 		if clr != c.Wild {
// 			if p.PropertyRents[clr] > 0 {
// 				return true
// 			}
// 		} else {
// 			for _, pc := range p.PropertyCards[c.Wild] {
// 				if pc.GetMoneyValue() > 0 {
// 					return true
// 				}
// 			}
// 		}
// 	}
// 	return false
// }

// func (p *Player) AddToBank(card c.Card) bool {
// 	switch card.GetType() {
// 	case c.CardTypeMoney, c.CardTypeAction, c.CardTypeRent:
// 		p.BankCards = append(p.BankCards, card)
// 		return true
// 	default:
// 		return false
// 	}
// }

// func (p *Player) AddToHand(card c.Card) {
// 	p.Mutex.Lock()
// 	p.HandCards[card.GetID()] = card
// 	p.Mutex.Unlock()
// }

// func (p *Player) GetFromHand(cardID uint8) (c.Card, bool) {
// 	if p.HasInHand(cardID) {
// 		card := p.HandCards[cardID]
// 		delete(p.HandCards, cardID)
// 		return card, true
// 	}
// 	return nil, false
// }

// func (p *Player) GetProperty(cardID uint8) (card c.Card, ok bool, from c.PropertyColor, commit func()) {
// 	// return func() is used to commit, bool to make sure only one time commit
// 	committed := false
// 	for i, card := range p.PropertyCards[c.Wild] {
// 		if card.GetID() == cardID {
// 			return card, true, c.Wild, func() {
// 				if !committed {
// 					p.PropertyCards[c.Wild] = slices.Delete(p.PropertyCards[c.Wild], i, i+1)
// 					p.UpdateRents(c.Wild)
// 					committed = true
// 				}
// 			}
// 		}
// 	}

// 	// the following will only work if card ids are assigned in sequence
// 	for color, cards := range p.PropertyCards {
// 		if len(cards) > 0 && cardID <= cards[len(cards)-1].GetID() || p.HasHouse(color) || p.HasHotel(color) {
// 			for i, card := range cards {
// 				if card.GetID() == cardID {
// 					if p.HasHouse(color) {
// 						if ac, ok := card.(*c.ActionCard); ok {
// 							if (ac.GetActionType() != c.ActionHouse || p.HasHotel(color)) && (ac.GetActionType() != c.ActionHotel) {
// 								return nil, false, "", func() {}
// 							}
// 						} else {
// 							return nil, false, "", func() {}
// 						}
// 					}
// 					return card, true, color, func() {
// 						if !committed {
// 							p.PropertyCards[color] = slices.Delete(p.PropertyCards[color], i, i+1)
// 							p.UpdateRents(color)
// 							committed = true
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}

// 	return nil, false, "", func() {}
// }

// func (p *Player) AddToProperty(color c.PropertyColor, card c.Card) bool {
// 	defer p.UpdateRents(color)
// 	switch card.GetType() {
// 	case c.CardTypeProperty:
// 		if p.HasCompleteSet(color) {
// 			return false
// 		}
// 		fmt.Println("Going to add property card to ", color)
// 		if pc, ok := card.(*c.PropertyCard); ok && (color == c.Wild || pc.GetColor() == color) {
// 			fmt.Println("done", ok)
// 			p.PropertyCards[color] = append(p.PropertyCards[color], card)
// 			return true
// 		}
// 	case c.CardTypeWildcard:
// 		if p.HasCompleteSet(color) {
// 			return false
// 		}
// 		fmt.Println("Going to add wildcard to ", color)
// 		if wc, ok := card.(*c.Wildcard); ok && (color == c.Wild || slices.Contains(wc.GetColors(), color) || slices.Contains(wc.GetColors(), c.Wild)) {
// 			fmt.Println("done", ok)
// 			p.PropertyCards[color] = append(p.PropertyCards[color], card)
// 			return true
// 		}

// 	case c.CardTypeAction:
// 		fmt.Println("Going to add action card to ", color)
// 		if ac, ok := card.(*c.ActionCard); ok && !IsARailroadOrUtility(color) {
// 			if color == c.Wild {
// 				fmt.Println("done", ok)
// 				p.PropertyCards[color] = append(p.PropertyCards[color], card)
// 				return true
// 			}
// 			if ac.GetActionType() == c.ActionHouse && p.HasCompleteSet(color) && !p.HasHouse(color) {
// 				fmt.Println("done", ok)
// 				p.PropertyCards[color] = append(p.PropertyCards[color], card)
// 				return true
// 			}
// 			if ac.GetActionType() == c.ActionHotel && p.HasHouse(color) && !p.HasHotel(color) {
// 				fmt.Println("done", ok)
// 				p.PropertyCards[color] = append(p.PropertyCards[color], card)
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// func (p *Player) UpdateRents(color c.PropertyColor) {
// 	rent, completion := p.GetRentOnColor(color)
// 	p.PropertyRents[color] = rent
// 	p.PropertyCompletion[color] = completion
// }

// func (p *Player) IsHandEmpty() bool {
// 	return len(p.HandCards) == 0
// }

// func (p *Player) GetHandCount() uint8 {
// 	return uint8(len(p.HandCards))
// }

// func (p *Player) HasInHand(cardID uint8) bool {
// 	_, exists := p.HandCards[cardID]
// 	return exists
// }

// func (p *Player) HasCompleteSet(color c.PropertyColor) bool {
// 	if color == c.Wild {
// 		return false
// 	}
// 	if cards, exists := p.PropertyCards[color]; exists {
// 		required := c.GetColorSetRequirement(color)
// 		// >= checked because house/hotel can be there
// 		return len(cards) >= required
// 	}
// 	return false
// }
// func (p *Player) HasHouse(color c.PropertyColor) bool {
// 	if color == c.Wild {
// 		return false
// 	}
// 	if cards, exists := p.PropertyCards[color]; exists {
// 		required := c.GetColorSetRequirement(color)
// 		return len(cards) >= required+1
// 	}
// 	return false
// }
// func (p *Player) HasHotel(color c.PropertyColor) bool {
// 	if color == c.Wild {
// 		return false
// 	}
// 	if cards, exists := p.PropertyCards[color]; exists {
// 		required := c.GetColorSetRequirement(color)
// 		return len(cards) >= required+2
// 	}
// 	return false
// }

// func (p *Player) HasProperty(color c.PropertyColor) bool {
// 	return len(p.PropertyCards[color]) != 0
// }

// func (p *Player) GetCardForPayment(cardID uint8) (c.Card, bool) {
// 	index := slices.IndexFunc(p.BankCards, func(b c.Card) bool { return b.GetID() == cardID })
// 	if index != -1 {
// 		card := p.BankCards[index]
// 		p.BankCards = slices.Delete(p.BankCards, index, index+1)
// 		return card, true
// 	}
// 	if card, ok, _, commit := p.GetProperty(cardID); ok {
// 		if card.GetMoneyValue() > 0 {
// 			commit()
// 			return card, true
// 		}
// 	}
// 	return nil, false
// }

// func (p *Player) HasAnyCompleteSet() bool {
// 	return slices.ContainsFunc(c.GetAllColors(), p.HasCompleteSet)
// }

// func (p *Player) HasAnyIncompleteSet() bool {
// 	for clr, cards := range p.PropertyCards {
// 		if len(cards) > 0 && (clr == c.Wild || !p.HasCompleteSet(clr)) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (p *Player) TransferPropertySet(color c.PropertyColor, target *Player) bool {
// 	// have to see what house / hotel do here: done, ok now
// 	// confirmed, house/hotel are included
// 	if p.HasCompleteSet(color) {
// 		for _, c := range p.PropertyCards[color] {
// 			target.AddToProperty(color, c)
// 		}
// 		// target.PropertyCards[color] = append(target.PropertyCards[color], p.PropertyCards[color]...)
// 		delete(p.PropertyCards, color)
// 		return true
// 	}

// 	return false
// }

// func (p *Player) GetRentOnColor(color c.PropertyColor) (c.MoneyValue, uint8) {
// 	if color == c.Wild {
// 		return c.MoneyZero, 0
// 	}
// 	count := len(p.PropertyCards[color])
// 	if p.HasHotel(color) {
// 		return c.GetRentValuesForColor(color)[count-3] + HotelRent, 3
// 	} else if p.HasHouse(color) {
// 		return c.GetRentValuesForColor(color)[count-2] + HouseRent, 2
// 	} else if count > 0 {
// 		if p.HasCompleteSet(color) {
// 			return c.GetRentValuesForColor(color)[count-1], 1
// 		} else {
// 			return c.GetRentValuesForColor(color)[count-1], 0
// 		}
// 	} else {
// 		// find in wild
// 		for _, card := range p.PropertyCards[c.Wild] {
// 			switch card.GetType() {
// 			case c.CardTypeProperty:
// 				if pc, ok := card.(*c.PropertyCard); ok && pc.GetColor() == color {
// 					return pc.GetRentValues()[0], 0
// 				}
// 			case c.CardTypeWildcard:
// 				if wc, ok := card.(*c.Wildcard); ok && slices.Contains(wc.GetColors(), color) {
// 					return c.GetRentValuesForColor(color)[0], 0
// 				}
// 			}
// 		}
// 	}
// 	return c.MoneyZero, 0
// }

// func (p *Player) HasRequiredPropertySets() bool {
// 	count := 0
// 	for _, k := range p.PropertyCompletion {
// 		if k > 0 {
// 			count++
// 		}
// 	}
// 	return count >= RequiredPropertySets
// }

// -------

// func (p *Player) TransferPropertyIDBased(card c.Card, target *Player, allowFromCompleteSet bool) bool {

// 	// transfered to wild pile so they can arrange it by themselves
// 	switch v := card.(type) {
// 	case *c.PropertyCard:
// 		// Remove the PropertyCard from its color set
// 		colorSet := v.GetColor()
// 		cards := p.PropertyCards[colorSet]
// 		for i, card := range cards {
// 			if card.GetID() == v.GetID() {
// 				// found but cannot transfer since it was part of complete set
// 				if !allowFromCompleteSet && p.HasCompleteSet(colorSet) {
// 					return false
// 				}
// 				// Transfer to target
// 				target.AddToProperty(c.Wild, card)
// 				// target.PropertyCards[c.Wild] = append(target.PropertyCards[colorSet], card)
// 				// Remove from current player
// 				p.PropertyCards[colorSet] = append(cards[:i], cards[i+1:]...)
// 				p.UpdateRents(colorSet)
// 				return true
// 			}
// 		}
// 	case *c.Wildcard:
// 		// Remove the Wildcard from the appropriate color set
// 		for _, color := range v.GetColors() {
// 			cards := p.PropertyCards[color]
// 			for i, card := range cards {
// 				if card.GetID() == v.GetID() {
// 					// found but cannot transfer since it was part of complete set
// 					if !allowFromCompleteSet && p.HasCompleteSet(color) {
// 						return false
// 					}
// 					target.AddToProperty(c.Wild, card)
// 					// target.PropertyCards[c.Wild] = append(target.PropertyCards[color], card)
// 					p.PropertyCards[color] = append(cards[:i], cards[i+1:]...)
// 					p.UpdateRents(color)
// 					return true
// 				}
// 			}
// 		}
// 	}

// 	// If not found in any color set, check wild set
// 	wildCards := p.PropertyCards[c.Wild]
// 	for i, wc := range wildCards {
// 		if wc.GetID() == card.GetID() {
// 			// no check added here since in wild pile,
// 			// there can only be properties, wilds and house/hotels
// 			// and all are transferable
// 			target.AddToProperty(c.Wild, card)
// 			// target.PropertyCards[c.Wild] = append(target.PropertyCards[c.Wild], card)
// 			p.PropertyCards[c.Wild] = append(wildCards[:i], wildCards[i+1:]...)
// 			p.UpdateRents(c.Wild)
// 			return true
// 		}
// 	}

// 	return false
// }

// !!!!!!!!!!!!!!!!!!!!!!
// HERE LIES A PROBLEM FROM WHERE? TO TRANSFER
// func (p *Player) TransferProperty(color c.PropertyColor, target *Player) bool {
// 	// Try to transfer from organized set first
// 	if props, ok := p.PropertyCards[color]; ok && len(props) > 0 {
// 		card := props[0]
// 		target.AddToProperty(color, card)
// 		// target.PropertyCards[color] = append(target.PropertyCards[color], card)
// 		p.PropertyCards[color] = props[1:]
// 		p.UpdateRents(color)
// 		return true
// 	}

// 	// Fallback: check wild pile for any wildcard that supports this color
// 	for i, card := range p.PropertyCards[c.Wild] {
// 		if wildcard, ok := card.(*c.Wildcard); ok {
// 			for _, clr := range wildcard.GetColors() {
// 				if clr == color {
// 					// Transfer this wildcard
// 					target.AddToProperty(c.Wild, wildcard)
// 					// target.PropertyCards[c.Wild] = append(target.PropertyCards[c.Wild], wildcard)
// 					p.PropertyCards[c.Wild] = append(
// 						p.PropertyCards[c.Wild][:i],
// 						p.PropertyCards[c.Wild][i+1:]...,
// 					)
// 					p.UpdateRents(c.Wild)
// 					return true
// 				}
// 			}
// 		}
// 	}

// 	// Property not found
// 	return false
// }

// func (p *Player) GetBankSum() uint8 {
// 	var sum uint8 = 0
// 	for _, bc := range p.BankCards {
// 		sum += uint8(bc.GetMoneyValue())
// 	}
// 	return sum
// }

// func (p *Player) CheckPlayerHasCardsPlaced(cards []c.Card) bool {

// 	// future plan is to permanently save cardMap
// 	// there is a check to be added that if a property has house/hotel than cards beneath can be used as money

// 	cardMap := make(map[uint8]bool)
// 	for _, card := range p.BankCards {
// 		cardMap[card.GetID()] = true
// 	}
// 	for _, pcs := range p.PropertyCards {
// 		for _, card := range pcs {
// 			cardMap[card.GetID()] = true
// 		}
// 	}
// 	for _, card := range cards {
// 		if !cardMap[card.GetID()] {
// 			return false
// 		}
// 	}
// 	return true
// }

// func (p *Player) TransferCardsAndPlace(target *Player, moneyCards []c.Card) bool {
// 	// placing in wild because player will get to arrange them in their turn
// 	for _, card := range moneyCards {
// 		switch v := card.(type) {
// 		case *c.MoneyCard, *c.ActionCard:
// 			// Add to target's bank
// 			target.BankCards = append(target.BankCards, card)

// 			// Remove from payer's bank
// 			for i, b := range p.BankCards {
// 				if b == v {
// 					p.BankCards = slices.Delete(p.BankCards, i, i+1)
// 					break
// 				}
// 			}

// 		case *c.PropertyCard, *c.Wildcard:
// 			// Add to target's generic Wild bucket (unorganized holding area)
// 			target.PropertyCards[c.Wild] = append(target.PropertyCards[c.Wild], v)

// 			// Remove from all property piles (including c.Wild)
// 			found := false
// 			for color, cards := range p.PropertyCards {
// 				for i, prop := range cards {
// 					if prop == v {
// 						p.PropertyCards[color] = append(cards[:i], cards[i+1:]...)
// 						found = true
// 						break
// 					}
// 				}
// 				if found {
// 					break
// 				}
// 			}

// 		}

// 		// TODO: Incomplete
// 	}
// 	return false
// }

// func (p *Player) TransferMoney(target *Player, amount c.MoneyValue, moneyCards []c.Card) bool {

// 	if !p.CheckPlayerHasCardsPlaced(moneyCards) {
// 		return false
// 	}
// 	var total uint8
// 	for _, card := range moneyCards {
// 		total += uint8(card.GetMoneyValue())
// 	}

// 	if total >= uint8(amount) {
// 		p.TransferCardsAndPlace(target, moneyCards)
// 		return true
// 	}

// 	return false
// }
