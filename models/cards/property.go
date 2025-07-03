package cards

type PropertyColor string

const (
	DarkBlue  PropertyColor = "Dark Blue"
	Brown     PropertyColor = "Brown"
	Green     PropertyColor = "Green"
	Red       PropertyColor = "Red"
	Yellow    PropertyColor = "Yellow"
	Orange    PropertyColor = "Orange"
	Pink      PropertyColor = "Pink"
	LightBlue PropertyColor = "Light Blue"
	Railroad  PropertyColor = "Railroad"
	Utility   PropertyColor = "Utility"
	Wild      PropertyColor = "Wild"
)

type PropertyCardInfo struct {
	Color      PropertyColor `json:"color"`
	RentValues []MoneyValue  `json:"rent_values"`
}

type PropertyCard struct {
	BaseCard
	PropertyCardInfo `json:"property_card_info"`
}

func (p *PropertyCard) GetType() CardType { return CardTypeProperty }
func (p *PropertyCard) GetColor() PropertyColor {
	return p.Color
}
func (p *PropertyCard) GetRentValues() []MoneyValue {
	return p.RentValues
}

func NewPropertyCard(name string, color PropertyColor, rentValues []MoneyValue, value MoneyValue) *PropertyCard {
	return &PropertyCard{
		BaseCard: BaseCard{
			Name:  name,
			Value: value,
			Type:  CardTypeProperty,
		},
		PropertyCardInfo: PropertyCardInfo{
			Color:      color,
			RentValues: rentValues,
		},
	}
}

func GetAllColors() []PropertyColor {
	return []PropertyColor{DarkBlue, Brown, Green, Red, Yellow, Orange, Pink, LightBlue, Railroad, Utility, Wild}
}

// TODO -- rent values incorrect
func GetRentValuesForColor(color PropertyColor) []MoneyValue {
	switch color {
	case DarkBlue:
		return []MoneyValue{Money3M, Money8M}
	case Brown:
		return []MoneyValue{Money1M, Money2M}
	case Green:
		return []MoneyValue{Money2M, Money4M, Money7M}
	case Yellow:
		return []MoneyValue{Money2M, Money4M, Money6M}
	case Red:
		return []MoneyValue{Money2M, Money3M, Money6M}
	case Orange:
		return []MoneyValue{Money1M, Money3M, Money5M}
	case Pink:
		return []MoneyValue{Money1M, Money2M, Money4M}
	case LightBlue:
		return []MoneyValue{Money1M, Money2M, Money3M}
	case Railroad:
		return []MoneyValue{Money1M, Money2M, Money3M, Money4M}
	case Utility:
		return []MoneyValue{Money1M, Money2M}
	default:
		return nil
	}
}

func NewBluePropertyCard(name string) *PropertyCard {
	return NewPropertyCard(name, DarkBlue, GetRentValuesForColor(DarkBlue), Money4M)
}
func NewBrownPropertyCard(name string) *PropertyCard {
	return NewPropertyCard(name, Brown, GetRentValuesForColor(Brown), Money3M)
}
func NewUtilityCard(name string) *PropertyCard {
	return NewPropertyCard(name, Utility, GetRentValuesForColor(Utility), Money2M)
}
func NewGreenPropertyCard(name string) *PropertyCard {
	return NewPropertyCard(name, Green, GetRentValuesForColor(Green), Money4M)
}
func NewYellowPropertyCard(name string) *PropertyCard {
	return NewPropertyCard(name, Yellow, GetRentValuesForColor(Yellow), Money3M)
}
func NewRedPropertyCard(name string) *PropertyCard {
	return NewPropertyCard(name, Red, GetRentValuesForColor(Red), Money4M)
}
func NewOrangePropertyCard(name string) *PropertyCard {
	return NewPropertyCard(name, Orange, GetRentValuesForColor(Orange), Money3M)
}
func NewPinkPropertyCard(name string) *PropertyCard {
	return NewPropertyCard(name, Pink, GetRentValuesForColor(Pink), Money4M)
}
func NewLightBluePropertyCard(name string) *PropertyCard {
	return NewPropertyCard(name, LightBlue, GetRentValuesForColor(LightBlue), Money3M)
}
func NewRailroadCard(name string) *PropertyCard {
	return NewPropertyCard(name, Railroad, GetRentValuesForColor(Railroad), Money5M)
}

func GetPropertyCardDeck() []Card {
	var deck []Card

	// Property Cards
	deck = append(deck,
		NewBluePropertyCard("Dark Blue 1"), NewBluePropertyCard("Dark Blue 2"),
		NewBrownPropertyCard("Brown 1"), NewBrownPropertyCard("Brown 2"),
		NewUtilityCard("Utility 1"), NewUtilityCard("Utility 2"),
		NewGreenPropertyCard("Green 1"), NewGreenPropertyCard("Green 2"), NewGreenPropertyCard("Green 3"),
		NewYellowPropertyCard("Yellow 1"), NewYellowPropertyCard("Yellow 2"), NewYellowPropertyCard("Yellow 3"),
		NewRedPropertyCard("Red 1"), NewRedPropertyCard("Red 2"), NewRedPropertyCard("Red 3"),
		NewOrangePropertyCard("Orange 1"), NewOrangePropertyCard("Orange 2"), NewOrangePropertyCard("Orange 3"),
		NewPinkPropertyCard("Pink 1"), NewPinkPropertyCard("Pink 2"), NewPinkPropertyCard("Pink 3"),
		NewLightBluePropertyCard("Light Blue 1"), NewLightBluePropertyCard("Light Blue 2"), NewLightBluePropertyCard("Light Blue 3"),
		NewRailroadCard("Railroad 1"), NewRailroadCard("Railroad 2"), NewRailroadCard("Railroad 3"), NewRailroadCard("Railroad 4"),
	)

	return deck
}
