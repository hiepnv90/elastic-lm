package common

import (
	"fmt"
	"math/big"
)

type Token struct {
	Amount   *big.Int
	Symbol   string
	Decimals int
}

func (t Token) String() string {
	return fmt.Sprintf("%s-%s", FormatAmount(t.Amount, t.Decimals, 5), t.Symbol)
}

func (t Token) Equal(o Token) bool {
	return t.Amount.Cmp(o.Amount) == 0 && t.Symbol == o.Symbol && t.Decimals == o.Decimals
}

func (t Token) IsStable() bool {
	return t.Symbol == "USDT" ||
		t.Symbol == "USDC" ||
		t.Symbol == "DAI" ||
		t.Symbol == "BUSD"
}

func (t Token) FormatAmount(precision int) string {
	return FormatAmount(t.Amount, t.Decimals, precision)
}
