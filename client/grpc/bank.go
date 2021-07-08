package grpc

import (
	"context"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// GetBankQueryClient returns a object of queryClient.
func (c *Client) GetBankQueryClient() banktypes.QueryClient {
	return banktypes.NewQueryClient(c)
}

// GetBalance returns balance of a given account for staking denom.
func (c *Client) GetBalance(ctx context.Context, address string, denom string) (*sdktypes.Coin, error) {
	bankClient := c.GetBankQueryClient()

	req := banktypes.QueryBalanceRequest{
		Address: address,
		Denom:   denom,
	}

	resp, err := bankClient.Balance(ctx, &req)
	if err != nil {
		return nil, err
	}

	return resp.GetBalance(), nil
}

func (c *Client) GetAllBalances(ctx context.Context, address string) (sdktypes.Coins, error) {
	bankClient := c.GetBankQueryClient()

	req := banktypes.QueryAllBalancesRequest{
		Address: address,
	}

	resp, err := bankClient.AllBalances(ctx, &req)
	if err != nil {
		return nil, err
	}
	return resp.GetBalances(), nil
}
