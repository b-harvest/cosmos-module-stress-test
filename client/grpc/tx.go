package grpc

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/tx"
)

// GetTxClient returns an object of ServiceClient.
func (c *Client) GetTxClient() tx.ServiceClient {
	return tx.NewServiceClient(c)
}

// BroadcastTx broadcasts transaction.
func (c *Client) BroadcastTx(ctx context.Context, txBytes []byte) (*tx.BroadcastTxResponse, error) {
	client := c.GetTxClient()

	req := &tx.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC, // should use async mode for the stress testing
	}

	return client.BroadcastTx(ctx, req)
}
