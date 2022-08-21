package graphql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var baseURL = "https://api.thegraph.com/subgraphs/name/kybernetwork/kyberswap-elastic-matic"

func TestGetPositions(t *testing.T) {
	client := New(baseURL, nil)

	_, err := client.GetPositions([]string{"799"})
	assert.NoError(t, err)
}
