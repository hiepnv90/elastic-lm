package common

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
)

type RoundType int

const (
	RoundTypeFloor RoundType = iota
	RoundTypeCeiling
)

var (
	Big0       = big.NewInt(0)
	Big1       = big.NewInt(1)
	Big2       = big.NewInt(2)
	MaxUint256 = new(big.Int).Sub(new(big.Int).Exp(Big2, big.NewInt(256), nil), Big1)
)

func NewBigIntFromString(s string, base int) *big.Int {
	b, ok := new(big.Int).SetString(s, base)
	if !ok {
		panic(fmt.Sprintf("fail to parse string to big.Int: s=%s base=%d", s, base))
	}
	return b
}

func NewBigIntFromHex(s string) *big.Int {
	return NewBigIntFromString(s, 16)
}

func BigAdd(a *big.Int, b *big.Int) *big.Int {
	return new(big.Int).Add(a, b)
}

func BigSub(a *big.Int, b *big.Int) *big.Int {
	return new(big.Int).Sub(a, b)
}

func BigMul(a *big.Int, b *big.Int) *big.Int {
	return new(big.Int).Mul(a, b)
}

func BigDiv(a *big.Int, b *big.Int) *big.Int {
	return new(big.Int).Div(a, b)
}

func BigAnd(a *big.Int, b *big.Int) *big.Int {
	return new(big.Int).And(a, b)
}

func BigIsZero(a *big.Int) bool {
	return a.Cmp(Big0) == 0
}

func BigPowerOf2(exp int64) *big.Int {
	return new(big.Int).Exp(Big2, big.NewInt(exp), nil)
}

func BigShiftLeft(a *big.Int, shift int64) *big.Int {
	return BigMul(a, BigPowerOf2(shift))
}

func BigShiftRight(a *big.Int, shift int64) *big.Int {
	return BigDiv(a, BigPowerOf2(shift))
}

func BigMod(a *big.Int, b *big.Int) *big.Int {
	return new(big.Int).Mod(a, b)
}

func BigNeg(a *big.Int) *big.Int {
	return new(big.Int).Neg(a)
}

func BigExp(a *big.Int, exp int64) *big.Int {
	return new(big.Int).Exp(a, big.NewInt(exp), nil)
}

func BigAbs(a *big.Int) *big.Int {
	return new(big.Int).Abs(a)
}

func RoundAmount(
	amount *big.Int, decimals int, precision int, roundType RoundType,
) *big.Int {
	if precision >= decimals {
		return amount
	}

	factor := BigExp(big.NewInt(10), int64(decimals-precision))
	round := BigDiv(amount, factor)
	if roundType == RoundTypeCeiling && !BigIsZero(BigMod(amount, factor)) {
		round = BigAdd(round, Big1)
	}

	return BigMul(round, factor)
}

func FormatAmount(amount *big.Int, decimals int, precision int) string {
	if precision < 0 {
		amount = RoundAmount(amount, decimals, precision, RoundTypeFloor)
		precision = 0
	} else if precision > decimals {
		precision = decimals
	}

	factor := BigExp(big.NewInt(10), int64(decimals))
	prec := BigExp(big.NewInt(10), int64(decimals-precision))
	return fmt.Sprintf(
		"%s.%0"+strconv.Itoa(precision)+"s",
		BigDiv(amount, factor).String(),
		BigDiv(BigMod(amount, factor), prec).String(),
	)
}

func FloatIsZero(f float64) bool {
	return math.Abs(f) < 1e10
}
