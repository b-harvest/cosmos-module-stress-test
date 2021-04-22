package grpc

import (
	"context"
	"fmt"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// GetBankQueryClient returns a object of queryClient
func (c *Client) GetBankQueryClient() banktypes.QueryClient {
	return banktypes.NewQueryClient(c)
}

// GetBalance returns balance of a given account for staking denom
func (c *Client) GetBalance(ctx context.Context, denom, address string) (*sdktypes.Coin, error) {
	bankClient := c.GetBankQueryClient()

	qr := banktypes.QueryBalanceRequest{
		Address: address,
		Denom:   denom,
	}

	res, err := bankClient.Balance(ctx, &qr)
	if err != nil {
		return nil, err
	}

	return res.GetBalance(), nil
}

// GetAllBalances returns balances of a given account
func (c *Client) GetAllBalances(ctx context.Context, address string, pageLimit uint64) (sdktypes.Coins, error) {
	bankClient := c.GetBankQueryClient()

	qr := banktypes.QueryAllBalancesRequest{
		Address: address,
		Pagination: &sdkquery.PageRequest{
			Limit: pageLimit,
		},
	}

	res, err := bankClient.AllBalances(ctx, &qr)
	if err != nil {
		if IsNotFound(err) {
			return sdktypes.Coins{}, nil
		}
		return nil, fmt.Errorf("failed to get all balances : %v", err)
	}

	return res.GetBalances(), nil
}
