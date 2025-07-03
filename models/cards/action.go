package cards

type ActionCardType string

const (
	ActionDealBreaker ActionCardType = "Deal Breaker"
	ActionJustSayNo   ActionCardType = "Just Say No"
	ActionPassGo      ActionCardType = "Pass Go"
	ActionForcedDeal  ActionCardType = "Forced Deal"
	ActionSlyDeal     ActionCardType = "Sly Deal"
	ActionDebtCollect ActionCardType = "Debt Collector"
	ActionBirthday    ActionCardType = "It's My Birthday"
	ActionHouse       ActionCardType = "House"
	ActionHotel       ActionCardType = "Hotel"
	ActionDoubleRent  ActionCardType = "Double The Rent"
)

type ActionCard struct {
	BaseCard
	actionType ActionCardType
}

func (a *ActionCard) GetType() CardType             { return CardTypeAction }
func (a *ActionCard) GetActionType() ActionCardType { return a.actionType }

func NewActionCard(acType ActionCardType, value MoneyValue) *ActionCard {
	return &ActionCard{
		BaseCard: BaseCard{
			Name:  string(acType),
			Value: value,
			Type:  CardTypeAction,
		},
		actionType: acType,
	}
}

func NewDealBreaker() *ActionCard {
	return NewActionCard(ActionDealBreaker, Money5M)
}
func NewJustSayNo() *ActionCard {
	return NewActionCard(ActionJustSayNo, Money4M)
}
func NewPassGo() *ActionCard {
	return NewActionCard(ActionPassGo, Money1M)
}
func NewForcedDeal() *ActionCard {
	return NewActionCard(ActionForcedDeal, Money3M)
}
func NewSlyDeal() *ActionCard {
	return NewActionCard(ActionSlyDeal, Money3M)
}
func NewDebtCollect() *ActionCard {
	return NewActionCard(ActionDebtCollect, Money3M)
}
func NewBirthday() *ActionCard {
	return NewActionCard(ActionBirthday, Money2M)
}
func NewDoubleRent() *ActionCard {
	return NewActionCard(ActionDoubleRent, Money1M)
}
func NewHouse() *ActionCard {
	return NewActionCard(ActionHouse, Money3M)
}
func NewHotel() *ActionCard {
	return NewActionCard(ActionHotel, Money4M)
}

const (
	DealBreakerCount = 2
	JustSayNoCount   = 3
	PassGoCount      = 8
	ForcedDealCount  = 4
	SlyDealCount     = 3
	DebtCollectCount = 3
	BirthdayCount    = 3
	DoubleRentCount  = 2
	HouseCount       = 3
	HotelCount       = 3
)

var newActionCardFuncs = []struct {
	count int
	newFn func() Card
}{
	{DealBreakerCount, func() Card { return NewDealBreaker() }},
	{JustSayNoCount, func() Card { return NewJustSayNo() }},
	{PassGoCount, func() Card { return NewPassGo() }},
	{ForcedDealCount, func() Card { return NewForcedDeal() }},
	{SlyDealCount, func() Card { return NewSlyDeal() }},
	{DebtCollectCount, func() Card { return NewDebtCollect() }},
	{BirthdayCount, func() Card { return NewBirthday() }},
	{DoubleRentCount, func() Card { return NewDoubleRent() }},
	{HouseCount, func() Card { return NewHouse() }},
	{HotelCount, func() Card { return NewHotel() }},
}

func GetActionCardDeck() []Card {
	var deck []Card

	for _, ac := range newActionCardFuncs {
		for i := 0; i < ac.count; i++ {
			deck = append(deck, ac.newFn())
		}
	}

	return deck
}

func IsActionPassGo(act ActionCardType) bool {
	return act == ActionPassGo
}

func IsActionTargetedAtSinglePlayer(act ActionCardType) bool {
	switch act {
	case ActionForcedDeal, ActionSlyDeal, ActionDebtCollect, ActionDealBreaker:
		return true
	default:
		return false
	}
}
