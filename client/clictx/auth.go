package clictx

import (
	"fmt"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// GetAccount checks account type and returns account interface.
func (c *Client) GetAccount(address string) (sdkclient.Account, error) {
	accAddr, err := sdktypes.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	ar := authtypes.AccountRetriever{}

	acc, _, err := ar.GetAccountWithHeight(c.Context, accAddr)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

// GetTx queries for a single transaction by a hash string in hex format.
// An error is returned if the transaction does not exist or cannot be queried.
func (c *Client) GetTx(hash string) (*sdktypes.TxResponse, error) {
	txResponse, err := authclient.QueryTx(c.Context, hash) // use RPC under the hood
	if err != nil {
		return &sdktypes.TxResponse{}, fmt.Errorf("failed to query tx hash: %s", err)
	}

	if txResponse.Empty() {
		return &sdktypes.TxResponse{}, fmt.Errorf("tx hash has empty tx response: %s", err)
	}

	return txResponse, nil
}

// GetTxs queries for all the transactions in a block.
// Transactions are returned in the sdktypes.TxResponse format which internally contains an sdktypes.Tx.
func (c *Client) GetTxs(block *tmctypes.ResultBlock) ([]*sdktypes.TxResponse, error) {
	txResponses := make([]*sdktypes.TxResponse, len(block.Block.Txs))

	if len(block.Block.Txs) <= 0 {
		return txResponses, nil
	}

	for i, tx := range block.Block.Txs {
		txResponse, err := c.GetTx(fmt.Sprintf("%X", tx.Hash()))
		if err != nil {
			return nil, err
		}

		txResponses[i] = txResponse
	}

	return txResponses, nil
}
