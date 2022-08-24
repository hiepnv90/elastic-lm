package common

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundAmount(t *testing.T) {
	tests := []struct {
		amount    *big.Int
		decimals  int
		precision int
		roundType RoundType
		expected  *big.Int
	}{
		{
			amount:    big.NewInt(123456),
			decimals:  5,
			precision: 0,
			roundType: RoundTypeFloor,
			expected:  big.NewInt(100000),
		},
		{
			amount:    big.NewInt(123456),
			decimals:  5,
			precision: 1,
			roundType: RoundTypeCeiling,
			expected:  big.NewInt(130000),
		},
		{
			amount:    big.NewInt(-123456),
			decimals:  5,
			precision: 0,
			roundType: RoundTypeFloor,
			expected:  big.NewInt(-200000),
		},
		{
			amount:    big.NewInt(-123456),
			decimals:  5,
			precision: 2,
			roundType: RoundTypeCeiling,
			expected:  big.NewInt(-123000),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, RoundAmount(test.amount, test.decimals, test.precision, test.roundType))
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		amount    *big.Int
		decimals  int
		precision int
		expected  string
	}{
		{
			amount:    big.NewInt(123456),
			decimals:  5,
			precision: 0,
			expected:  "1.0",
		},
		{
			amount:    big.NewInt(123456),
			decimals:  5,
			precision: 2,
			expected:  "1.23",
		},
		{
			amount:    big.NewInt(-123456),
			decimals:  5,
			precision: 0,
			expected:  "-1.0",
		},
		{
			amount:    big.NewInt(-8145849),
			decimals:  6,
			precision: 5,
			expected:  "-8.14584",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, FormatAmount(test.amount, test.decimals, test.precision))
	}
}
