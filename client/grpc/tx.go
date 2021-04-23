package grpc

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/types/tx"
)

// GetTxClient returns an object of ServiceClient.
func (c *Client) GetTxClient() tx.ServiceClient {
	return tx.NewServiceClient(c)
}

// BroadcastTx broadcasts transaction.
func (c *Client) BroadcastTx(txBytes []byte) (*tx.BroadcastTxResponse, error) {
	txClient := c.GetTxClient()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	qr := &tx.BroadcastTxRequest{
		TxBytes: txBytes, // raw transaction
		Mode:    tx.BroadcastMode_BROADCAST_MODE_ASYNC,
	}

	return txClient.BroadcastTx(ctx, qr)
}
