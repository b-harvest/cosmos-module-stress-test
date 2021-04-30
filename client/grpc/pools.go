package grpc

import (
	"context"

	sdkquery "github.com/cosmos/cosmos-sdk/types/query"

	liquiditytypes "github.com/tendermint/liquidity/x/liquidity/types"
)

// GetLiquidityQueryClient returns a object of queryClient
func (c *Client) GetLiquidityQueryClient() liquiditytypes.QueryClient {
	return liquiditytypes.NewQueryClient(c)
}

// GetPool returns pool information.
func (c *Client) GetPool(ctx context.Context, poolId uint64) (liquiditytypes.Pool, error) {
	client := c.GetLiquidityQueryClient()

	req := liquiditytypes.QueryLiquidityPoolRequest{
		PoolId: poolId,
	}

	resp, err := client.LiquidityPool(ctx, &req)
	if err != nil {
		return liquiditytypes.Pool{}, err
	}

	return resp.GetPool(), nil
}

// GetAllPools returns all existing pools.
func (c *Client) GetAllPools(ctx context.Context) (liquiditytypes.Pools, error) {
	client := c.GetLiquidityQueryClient()

	req := liquiditytypes.QueryLiquidityPoolsRequest{
		Pagination: &sdkquery.PageRequest{},
	}

	resp, err := client.LiquidityPools(ctx, &req)
	if err != nil {
		return liquiditytypes.Pools{}, err
	}

	return resp.GetPools(), nil
}
