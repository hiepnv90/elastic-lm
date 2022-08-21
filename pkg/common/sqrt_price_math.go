package common

import (
	"math/big"
)

func CalculateAmount0(lowerSqrtPrice *big.Int, upperSqrtPrice *big.Int, liquidity *big.Int) *big.Int {
	numerator1 := BigShiftLeft(liquidity, 96)
	numerator2 := BigAbs(BigSub(upperSqrtPrice, lowerSqrtPrice))
	return BigDiv(
		BigMul(numerator1, numerator2),
		BigMul(upperSqrtPrice, lowerSqrtPrice),
	)
}

func CalculateAmount1(lowerSqrtPrice *big.Int, upperSqrtPrice *big.Int, liquidity *big.Int) *big.Int {
	return BigShiftRight(BigMul(liquidity, BigAbs(BigSub(upperSqrtPrice, lowerSqrtPrice))), 96)
}
