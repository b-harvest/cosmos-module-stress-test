package clictx

import (
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
