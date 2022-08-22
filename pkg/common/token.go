package common

import (
	"fmt"
	"math/big"
	"strings"
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
	symbol := strings.ToUpper(t.Symbol)
	return symbol == "USDT" ||
		symbol == "USDC" ||
		symbol == "DAI" ||
		symbol == "BUSD" ||
		symbol == "MUSD" ||
		symbol == "USDK" ||
		symbol == "MIMATIC"
}

func (t Token) RoundAmount(precision int, roundType RoundType) *big.Int {
	return RoundAmount(t.Amount, t.Decimals, precision, roundType)
}

func (t Token) FormatAmount(precision int) string {
	return FormatAmount(t.Amount, t.Decimals, precision)
}

func (t Token) NormalizedSymbol() string {
	symbol := strings.ToUpper(t.Symbol)
	switch symbol {
	case "STMATIC", "WMATIC":
		return "MATIC"
	case "WBTC":
		return "BTC"
	case "WETH":
		return "ETH"
	default:
		return symbol
	}
}

func (t Token) GetBinancePerpetualSymbol() string {
	return t.NormalizedSymbol() + "USDT"
}
