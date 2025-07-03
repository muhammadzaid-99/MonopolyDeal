package cards

type RentCard struct {
	BaseCard
	Colors []PropertyColor `json:"colors"`
}

func (r *RentCard) GetType() CardType { return CardTypeRent }

func (r *RentCard) GetColors() []PropertyColor {
	return r.Colors
}

func NewRentCard(name string, colors []PropertyColor, value MoneyValue) *RentCard {
	return &RentCard{
		BaseCard: BaseCard{
			Name:  name,
			Value: value,
			Type:  CardTypeRent,
		},
		Colors: colors,
	}
}

func NewDarkBlueGreenRentCard() *RentCard {
	return NewRentCard("Rent - Dark Blue & Green", []PropertyColor{DarkBlue, Green}, Money2M)
}
func NewRedYellowRentCard() *RentCard {
	return NewRentCard("Rent - Red & Yellow", []PropertyColor{Red, Yellow}, Money2M)
}
func NewPinkOrangeRentCard() *RentCard {
	return NewRentCard("Rent - Pink & Orange", []PropertyColor{Pink, Orange}, Money2M)
}
func NewLightBlueBrownRentCard() *RentCard {
	return NewRentCard("Rent - Light Blue & Brown", []PropertyColor{LightBlue, Brown}, Money2M)
}
func NewRailroadUtilityRentCard() *RentCard {
	return NewRentCard("Rent - Railroad & Utility", []PropertyColor{Railroad, Utility}, Money2M)
}
func NewWildRentCard() *RentCard {
	return NewRentCard("Rent - Wild", []PropertyColor{Wild}, Money2M)
}

func GetRentCardDeck() []Card {
	var deck []Card

	for _, rent := range []struct {
		newFn func() Card
		count int
	}{
		{func() Card { return NewDarkBlueGreenRentCard() }, 2},
		{func() Card { return NewRedYellowRentCard() }, 2},
		{func() Card { return NewPinkOrangeRentCard() }, 2},
		{func() Card { return NewLightBlueBrownRentCard() }, 2},
		{func() Card { return NewRailroadUtilityRentCard() }, 2},
		{func() Card { return NewWildRentCard() }, 3},
	} {
		for i := 0; i < rent.count; i++ {
			deck = append(deck, rent.newFn())
		}
	}
	return deck
}
