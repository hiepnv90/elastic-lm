package common

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateAmount0(t *testing.T) {
	tests := []struct {
		lowerSqrtPrice *big.Int
		upperSqrtPrice *big.Int
		liquidity      *big.Int
		expected       *big.Int
	}{
		{
			lowerSqrtPrice: GetSqrtRatioAtTick(-276310),
			upperSqrtPrice: GetSqrtRatioAtTick(-276300),
			liquidity:      BigExp(big.NewInt(10), 14),
			expected:       NewBigIntFromString("49949961958869841", 10),
		},
		{
			lowerSqrtPrice: GetSqrtRatioAtTick(-15500),
			upperSqrtPrice: GetSqrtRatioAtTick(-14600),
			liquidity:      NewBigIntFromString("4521273292232113180183", 10),
			expected:       NewBigIntFromString("431795842829084192009", 10),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, CalculateAmount0(test.lowerSqrtPrice, test.upperSqrtPrice, test.liquidity))
	}
}

func TestCalculateAmount1(t *testing.T) {
	tests := []struct {
		lowerSqrtPrice *big.Int
		upperSqrtPrice *big.Int
		liquidity      *big.Int
		expected       *big.Int
	}{
		{
			lowerSqrtPrice: GetSqrtRatioAtTick(-276310),
			upperSqrtPrice: GetSqrtRatioAtTick(-276300),
			liquidity:      BigExp(big.NewInt(10), 14),
			expected:       NewBigIntFromString("50045", 10),
		},
		{
			lowerSqrtPrice: GetSqrtRatioAtTick(-16600),
			upperSqrtPrice: GetSqrtRatioAtTick(-15500),
			liquidity:      NewBigIntFromString("4521273292232113180183", 10),
			expected:       NewBigIntFromString("111468606089896287952", 10),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, CalculateAmount1(test.lowerSqrtPrice, test.upperSqrtPrice, test.liquidity))
	}
}
