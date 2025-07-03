package cards

type Card interface {
	GetName() string
	GetMoneyValue() MoneyValue
	GetType() CardType
	GetID() uint8
	SetID(id uint8)
}

type BaseCard struct {
	Name  string     `json:"name"`
	ID    uint8      `json:"id"`
	Value MoneyValue `json:"value"`
	Type  CardType   `json:"card_type"`
}

func (b *BaseCard) GetName() string { return b.Name }
func (b *BaseCard) GetMoneyValue() MoneyValue {
	return b.Value
}

func (b *BaseCard) GetID() uint8 {
	return b.ID
}

func (b *BaseCard) SetID(id uint8) {
	b.ID = id
}

type CardType string

const (
	CardTypeMoney    CardType = "Money"
	CardTypeAction   CardType = "Action"
	CardTypeProperty CardType = "Property"
	CardTypeRent     CardType = "Rent"
	CardTypeWildcard CardType = "Wildcard"
)
