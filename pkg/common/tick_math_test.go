package common

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSqrtRatioAtTick(t *testing.T) {
	tests := []struct {
		tick     int
		expected *big.Int
	}{
		{
			tick:     0,
			expected: BigPowerOf2(96),
		},
		{
			tick:     -887272,
			expected: big.NewInt(4295128739),
		},
		{
			tick:     887272,
			expected: NewBigIntFromString("1461446703485210103287273052203988822378723970342", 10),
		},
		{
			tick:     1,
			expected: NewBigIntFromString("79232123823359799118286999568", 10),
		},
		{
			tick:     -1,
			expected: NewBigIntFromString("79224201403219477170569942574", 10),
		},
		{
			tick:     2,
			expected: NewBigIntFromString("79236085330515764027303304732", 10),
		},
		{
			tick:     3,
			expected: NewBigIntFromString("79240047035742135098198828268", 10),
		},
		{
			tick:     2559,
			expected: NewBigIntFromString("90041927759339286931870012045", 10),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, GetSqrtRatioAtTick(test.tick))
	}
}
