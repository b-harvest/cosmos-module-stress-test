package grpc

import (
	"context"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetAuthQueryClient returns a object of queryClient
func (c *Client) GetAuthQueryClient() authtypes.QueryClient {
	return authtypes.NewQueryClient(c)
}

// GetBaseAccountInfo returns base account information
func (c *Client) GetBaseAccountInfo(ctx context.Context, address string) (authtypes.BaseAccount, error) {
	authClient := c.GetAuthQueryClient()

	qr := authtypes.QueryAccountRequest{
		Address: address,
	}

	authRes, err := authClient.Account(ctx, &qr)
	if err != nil {
		return authtypes.BaseAccount{}, err
	}

	var acc authtypes.BaseAccount
	err = acc.Unmarshal(authRes.GetAccount().Value)
	if err != nil {
		return authtypes.BaseAccount{}, err
	}

	return acc, nil
}
