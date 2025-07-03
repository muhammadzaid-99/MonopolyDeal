package cards

import "encoding/json"

type MoneyValue uint8

const (
	MoneyZero MoneyValue = 0
	Money1M   MoneyValue = 1
	Money2M   MoneyValue = 2
	Money3M   MoneyValue = 3
	Money4M   MoneyValue = 4
	Money5M   MoneyValue = 5
	Money6M   MoneyValue = 6
	Money7M   MoneyValue = 7
	Money8M   MoneyValue = 8
	Money9M   MoneyValue = 9
	Money10M  MoneyValue = 10
)

type MoneyCard struct {
	BaseCard
}

func (m MoneyValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint8(m))
}

func (m *MoneyValue) UnmarshalJSON(data []byte) error {
	var val uint8
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	*m = MoneyValue(val)
	return nil
}

func (m *MoneyCard) GetType() CardType { return CardTypeMoney }

func NewMoneyCard(value MoneyValue) *MoneyCard {
	return &MoneyCard{
		BaseCard: BaseCard{
			Name:  "Money",
			Value: value,
			Type:  CardTypeMoney,
		},
	}
}

func GetMoneyCardDeck() []Card {
	var deck []Card

	for _, money := range []struct {
		value MoneyValue
		count int
	}{
		{Money1M, 6},
		{Money2M, 5},
		{Money3M, 3},
		{Money4M, 3},
		{Money5M, 2},
		{Money10M, 1},
	} {
		for i := 0; i < money.count; i++ {
			deck = append(deck, NewMoneyCard(money.value))
		}
	}
	return deck
}
