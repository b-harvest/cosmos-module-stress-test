package client

import (
	sdkclient "github.com/cosmos/cosmos-sdk/client"

	"github.com/b-harvest/liquidity-stress-test/client/clictx"
	"github.com/b-harvest/liquidity-stress-test/client/grpc"
	"github.com/b-harvest/liquidity-stress-test/client/rpc"
	"github.com/b-harvest/liquidity-stress-test/codec"
)

// Client is a wrapper for various clients.
type Client struct {
	CliCtx *clictx.Client
	RPC    *rpc.Client
	GRPC   *grpc.Client
}

// NewClient creates a new Client with the given configuration.
func NewClient(rpcURL string, grpcURL string) *Client {
	codec.SetCodec()

	rpcClient := rpc.NewClient(rpcURL, 5)
	grpcClient := grpc.NewClient(grpcURL, 5)
	cliCtx := clictx.NewClient(rpcURL, rpcClient.Client)

	return &Client{
		CliCtx: cliCtx,
		RPC:    rpcClient,
		GRPC:   grpcClient,
	}
}

// GetCLIContext returns client context for the network.
func (c *Client) GetCLIContext() sdkclient.Context {
	return c.CliCtx.Context
}

// GetRPCClient returns RPC client.
func (c *Client) GetRPCClient() *rpc.Client {
	return c.RPC
}

// GetGRPCClient returns GRPC client.
func (c *Client) GetGRPCClient() *grpc.Client {
	return c.GRPC
}
