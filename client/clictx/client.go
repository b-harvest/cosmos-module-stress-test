package clictx

import (
	"github.com/b-harvest/cosmos-module-stress-test/codec"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// Client wraps Cosmos SDK client context.
type Client struct {
	sdkclient.Context
}

// NewClient creates Cosmos SDK client.
func NewClient(rpcURL string, rpcClient rpcclient.Client) *Client {
	cliCtx := sdkclient.Context{}.
		WithNodeURI(rpcURL).
		WithClient(rpcClient).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithJSONMarshaler(codec.EncodingConfig.Marshaler).
		WithLegacyAmino(codec.EncodingConfig.Amino).
		WithTxConfig(codec.EncodingConfig.TxConfig).
		WithInterfaceRegistry(codec.EncodingConfig.InterfaceRegistry)

	return &Client{cliCtx}
}
