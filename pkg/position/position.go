package position

import (
	"fmt"
	"math/big"

	"github.com/hiepnv90/elastic-lm/pkg/common"
)

type Position struct {
	ID            string
	Liquidity     *big.Int
	TickLower     int
	TickUpper     int
	MaxAmount0    *big.Int
	MaxAmount1    *big.Int
	HedgedAmount0 *big.Int
	HedgedAmount1 *big.Int
	Token0        common.Token
	Token1        common.Token
}

func (p Position) String() string {
	return fmt.Sprintf(
		"PositionID=%s Liquidity=(%s, %s)",
		p.ID, p.Token0.String(), p.Token1.String(),
	)
}

func (p Position) Equal(o Position) bool {
	return p.Token0.Equal(o.Token0) && p.Token1.Equal(o.Token1)
}
