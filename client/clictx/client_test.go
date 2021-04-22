package clictx_test

import (
	"os"
	"testing"

	"github.com/b-harvest/liquidity-stress-test/client/clictx"
	"github.com/b-harvest/liquidity-stress-test/client/rpc"
	"github.com/b-harvest/liquidity-stress-test/codec"

	"github.com/test-go/testify/require"
)

var (
	c *clictx.Client

	rpcURL = "http://localhost:26657"
)

func TestMain(m *testing.M) {
	codec.SetCodec()

	c = clictx.NewClient(rpcURL, rpc.NewClient(rpcURL, 5))

	os.Exit(m.Run())
}

func TestGetAccount(t *testing.T) {
	address := ""
	acc, err := c.GetAccount(address)
	require.NoError(t, err)

	t.Log("pubkey :", acc.GetPubKey())
}
