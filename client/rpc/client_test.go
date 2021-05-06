package rpc_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/b-harvest/liquidity-stress-test/client/rpc"
	"github.com/b-harvest/liquidity-stress-test/codec"

	"github.com/test-go/testify/require"
)

var (
	c *rpc.Client

	rpcAddress = "http://localhost:26657"
)

func TestMain(m *testing.M) {
	codec.SetCodec()

	c, _ = rpc.NewClient(rpcAddress, 5)

	os.Exit(m.Run())
}

func TestGetNetworkChainID(t *testing.T) {
	chainID, err := c.GetNetworkChainID(context.Background())
	require.NoError(t, err)

	t.Log(chainID)
}

func TestGetBlockTime(t *testing.T) {
	ctx := context.Background()

	status, err := c.GetStatus(ctx)
	require.NoError(t, err)

	currentBlockHeight := status.SyncInfo.LatestBlockHeight
	currentBlockTime := status.SyncInfo.LatestBlockTime

	height := int64(currentBlockHeight - 1)
	prevBlock, err := c.Block(ctx, &height)
	require.NoError(t, err)

	prevBlockHeight := prevBlock.Block.Height
	prevBlockTime := prevBlock.Block.Time

	blockTime := currentBlockTime.Sub(prevBlockTime)

	fmt.Println(currentBlockHeight, currentBlockTime)
	fmt.Println(prevBlockHeight, prevBlockTime)
	fmt.Println(blockTime)
}
