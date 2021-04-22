package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAllBalances(t *testing.T) {
	ctx := context.Background()

	address := ""
	resp, err := c.GetAllBalances(ctx, address, 100)
	require.NoError(t, err)

	t.Log(resp)
}
