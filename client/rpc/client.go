package rpc

import (
	"context"
	"fmt"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpc "github.com/tendermint/tendermint/rpc/client/http"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Client wraps RPC client connection
type Client struct {
	rpcclient.Client
}

// NewClient creates RPC client
func NewClient(rpcURL string, timeout int64) *Client {
	rpcClient, err := rpc.NewWithTimeout(rpcURL, "/websocket", uint(timeout))
	if err != nil {
		panic(fmt.Errorf("failed to connect RPC client: %s", err))
	}

	return &Client{rpcClient}
}

// GetNetworkChainID returns network chain id
func (c *Client) GetNetworkChainID(ctx context.Context) (string, error) {
	status, err := c.Status(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get status: %v", err)
	}

	return status.NodeInfo.Network, nil
}

// GetStatus returns the status of the blockchain network
func (c *Client) GetStatus(ctx context.Context) (*tmctypes.ResultStatus, error) {
	return c.Status(ctx)
}

// GetBlock returns block information for the height
func (c *Client) GetBlock(ctx context.Context, height int64) (*tmctypes.ResultBlock, error) {
	return c.Block(ctx, &height)
}

// GetLatestBlockHeight returns the latest block height on the network
func (c *Client) GetLatestBlockHeight(ctx context.Context) (int64, error) {
	status, err := c.Status(ctx)
	if err != nil {
		return -1, err
	}

	return status.SyncInfo.LatestBlockHeight, nil
}