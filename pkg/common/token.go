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
	symbol := t.NormalizedSymbol()
	return symbol == "USDT" ||
		symbol == "USDC" ||
		symbol == "DAI" ||
		symbol == "BUSD" ||
		symbol == "MUSD" ||
		symbol == "USDK" ||
		symbol == "MIMATIC" ||
		symbol == "MAI" ||
		symbol == "MIM"
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
	case "USDT.E":
		return "USDT"
	case "DAI.E":
		return "DAI"
	case "USDC.E":
		return "USDC"
	case "STMATIC", "WMATIC":
		return "MATIC"
	case "WBTC", "WBTC.E":
		return "BTC"
	case "WETH", "WETH.E":
		return "ETH"
	case "WAVAX", "SAVAX":
		return "AVAX"
	case "LINK.E":
		return "LINK"
	case "AAVE.E":
		return "AAVE"
	case "MKNC":
		return "KNC"
	default:
		return symbol
	}
}

func (t Token) GetBinancePerpetualSymbol(quoteCurrency string) string {
	return t.NormalizedSymbol() + quoteCurrency
}
