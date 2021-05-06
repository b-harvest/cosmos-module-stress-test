package client_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/test-go/testify/require"

	"github.com/b-harvest/liquidity-stress-test/client"
)

var (
	c *client.Client

	rpcAddress  = "http://localhost:26657"
	grpcAddress = "localhost:9090"
)

func TestMain(m *testing.M) {
	c, _ = client.NewClient(rpcAddress, grpcAddress)

	os.Exit(m.Run())
}

func TestParseTxs(t *testing.T) {
	height := int64(28936) // debug issue

	block, err := c.RPC.GetBlock(&height)
	require.NoError(t, err)

	txs, err := c.CliCtx.GetTxs(block)
	require.NoError(t, err)

	for _, tx := range txs {
		fmt.Println("hash: ", tx)
	}
}
