package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPools(t *testing.T) {
	pools, err := c.GetAllPools(context.Background())
	require.NoError(t, err)

	for _, p := range pools {
		t.Log(p)
	}
}
