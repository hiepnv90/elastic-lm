package common

import (
	"math/big"
)

func ExtractLiquidity(currentTick int, tickLower int, tickUpper int, sqrtPrice *big.Int, liquidity *big.Int) (*big.Int, *big.Int) {
	if currentTick < tickLower {
		return CalculateAmount0(
			GetSqrtRatioAtTick(tickLower),
			GetSqrtRatioAtTick(tickUpper),
			liquidity,
		), Big0
	}

	if currentTick > tickUpper {
		return Big0, CalculateAmount1(
			GetSqrtRatioAtTick(tickLower),
			GetSqrtRatioAtTick(tickUpper),
			liquidity,
		)
	}

	amount0 := CalculateAmount0(sqrtPrice, GetSqrtRatioAtTick(tickUpper), liquidity)
	amount1 := CalculateAmount1(GetSqrtRatioAtTick(tickLower), sqrtPrice, liquidity)
	return amount0, amount1
}
