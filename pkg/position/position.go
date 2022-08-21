package position

import (
	"fmt"

	"github.com/hiepnv90/elastic-farm/pkg/common"
)

type Position struct {
	Token0 common.Token
	Token1 common.Token
}

func (p Position) String() string {
	return fmt.Sprintf("(%s, %s)", p.Token0.String(), p.Token1.String())
}

func (p Position) Equal(o Position) bool {
	return p.Token0.Equal(o.Token0) && p.Token1.Equal(o.Token1)
}
