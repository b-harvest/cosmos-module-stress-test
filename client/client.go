package client

import (
	sdkclient "github.com/cosmos/cosmos-sdk/client"

	"github.com/nodebreaker0-0/cosmos-module-stress-test/client/clictx"
	"github.com/nodebreaker0-0/cosmos-module-stress-test/client/grpc"
	"github.com/nodebreaker0-0/cosmos-module-stress-test/client/rpc"
	"github.com/nodebreaker0-0/cosmos-module-stress-test/codec"
)

// Client is a wrapper for various clients.
type Client struct {
	CliCtx *clictx.Client
	RPC    *rpc.Client
	GRPC   *grpc.Client
}

// NewClient creates a new Client with the given configuration.
func NewClient(rpcURL string, grpcURL string) (*Client, error) {
	codec.SetCodec()

	rpcClient, err := rpc.NewClient(rpcURL, 5)
	if err != nil {
		return &Client{}, err
	}

	grpcClient, err := grpc.NewClient(grpcURL, 5)
	if err != nil {
		return &Client{}, err
	}

	cliCtx := clictx.NewClient(rpcURL, rpcClient.Client)

	return &Client{
		CliCtx: cliCtx,
		RPC:    rpcClient,
		GRPC:   grpcClient,
	}, nil
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

// Stop defers the node stop execution to the RPC and GRPC clients.
func (c Client) Stop() error {
	err := c.RPC.Stop()
	if err != nil {
		return err
	}

	err = c.GRPC.Close()
	if err != nil {
		return err
	}
	return nil
}
