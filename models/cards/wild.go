package cards

type Wildcard struct {
	BaseCard
	Properties []PropertyCardInfo `json:"properties"`
}

func (w *Wildcard) GetType() CardType { return CardTypeWildcard }

// have to see if this is needed
// also depends on placement on pile
// so will do later what may be done
func (w *Wildcard) GetProperties() []PropertyCardInfo {
	return w.Properties
}

func (w *Wildcard) GetColors() []PropertyColor {
	colors := make([]PropertyColor, len(w.Properties))
	for i, p := range w.Properties {
		colors[i] = p.Color
	}
	return colors
}

func extractPropertyInfo(propertyColors []PropertyColor) []PropertyCardInfo {
	var propertyInfos []PropertyCardInfo
	for _, color := range propertyColors {
		propertyInfos = append(propertyInfos, PropertyCardInfo{
			Color:      color,
			RentValues: GetRentValuesForColor(color),
		})
	}
	return propertyInfos
}

func GetWildNameFromColors(propertyColors []PropertyColor) string {
	name := "Wild -"
	for _, color := range propertyColors {
		name += " " + string(color)
	}
	return name
}

func NewPropertyWildcard(propertyColors []PropertyColor, value MoneyValue) *Wildcard {
	return &Wildcard{
		BaseCard: BaseCard{
			Name:  GetWildNameFromColors(propertyColors),
			Value: value,
			Type:  CardTypeWildcard,
		},
		Properties: extractPropertyInfo(propertyColors),
	}
}

func GetWildCardDeck() []Card {
	var deck []Card

	deck = append(deck,
		NewPropertyWildcard([]PropertyColor{DarkBlue, Green}, Money4M),
		NewPropertyWildcard([]PropertyColor{Green, Railroad}, Money4M),
		NewPropertyWildcard([]PropertyColor{Utility, Railroad}, Money2M),
		NewPropertyWildcard([]PropertyColor{LightBlue, Railroad}, Money4M),
		NewPropertyWildcard([]PropertyColor{LightBlue, Brown}, Money1M),
		NewPropertyWildcard([]PropertyColor{Pink, Orange}, Money2M),
		NewPropertyWildcard([]PropertyColor{Pink, Orange}, Money2M),
		NewPropertyWildcard([]PropertyColor{Red, Yellow}, Money2M),
		NewPropertyWildcard([]PropertyColor{Red, Yellow}, Money2M),
		NewPropertyWildcard([]PropertyColor{Wild}, MoneyZero),
		NewPropertyWildcard([]PropertyColor{Wild}, MoneyZero),
	)

	return deck
}
