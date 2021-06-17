package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBalance(t *testing.T) {
	address := "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v"
	denom := "uatom"
	resp, err := c.GetBalance(context.Background(), address, denom)
	require.NoError(t, err)

	t.Log(resp)
}
