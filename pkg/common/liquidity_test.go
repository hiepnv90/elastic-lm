package common

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractLiquidity(t *testing.T) {
	tests := []struct {
		currentTick     int
		tickLower       int
		tickUpper       int
		sqrtPrice       *big.Int
		liquidity       *big.Int
		expectedAmount0 *big.Int
		expectedAmount1 *big.Int
	}{
		{
			currentTick:     -15500,
			tickLower:       -16600,
			tickUpper:       -14600,
			sqrtPrice:       GetSqrtRatioAtTick(-15500),
			liquidity:       NewBigIntFromString("4521273292232113180183", 10),
			expectedAmount0: NewBigIntFromString("431795842829084192009", 10),
			expectedAmount1: NewBigIntFromString("111468606089896287952", 10),
		},
	}

	for _, test := range tests {
		amount0, amount1 := ExtractLiquidity(test.currentTick, test.tickLower, test.tickUpper, test.sqrtPrice, test.liquidity)
		assert.Equal(t, test.expectedAmount0, amount0)
		assert.Equal(t, test.expectedAmount1, amount1)
	}
}
