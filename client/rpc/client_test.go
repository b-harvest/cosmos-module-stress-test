package rpc_test

import (
	"context"
	"os"
	"testing"

	"github.com/b-harvest/liquidity-stress-test/client/rpc"
	"github.com/b-harvest/liquidity-stress-test/codec"

	"github.com/test-go/testify/require"
)

var (
	c *rpc.Client

	rpcURL = "http://localhost:26657"
)

func TestMain(m *testing.M) {
	codec.SetCodec()

	c = rpc.NewClient(rpcURL, 5)

	os.Exit(m.Run())
}

func TestGetNetworkChainID(t *testing.T) {
	chainID, err := c.GetNetworkChainID(context.Background())
	require.NoError(t, err)

	t.Log(chainID)
}
