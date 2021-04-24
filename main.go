package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/b-harvest/liquidity-stress-test/client"
	"github.com/b-harvest/liquidity-stress-test/config"
	"github.com/b-harvest/liquidity-stress-test/transaction"
	"github.com/b-harvest/liquidity-stress-test/wallet"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

/*
	1. CreatePool은 MultiMsg로 한방에
	2. user1: Deposit, withdraw
	3. user2: Swap
*/

var (
	configPath = "./config.toml"

	gasLimit = uint64(100000000)
	fees     = sdktypes.NewCoins(sdktypes.NewCoin("stake", sdktypes.NewInt(0)))
	memo     = ""
)

func main() {
	// configure log level and use pretty logging
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("Liquidity Stress Testing Begins...")

	cfg, err := config.Read(configPath)
	if err != nil {
		log.Panic().Err(err).Msg("failed to read config file")
	}

	client := client.NewClient(cfg.RPC.Address, cfg.GRPC.Address)

	// alice creates a total number of 55 pairs of different pools
	alice(client)

	// bob requests deposits and withdraws
}

// alice creates every pair of coins exist in the network.
// The Gravity DEX testnet will have 11 coin types available.
// This will require to create a total number of 55 pairs of liquidity pools.
func alice(client *client.Client) error {
	ctx := context.Background()

	mnemonic := "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"

	chainID, err := client.RPC.GetNetworkChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain id: %s", err)
	}

	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(mnemonic, "")
	if err != nil {
		return fmt.Errorf("failed to retrieve account from mnemonic: %s", err)
	}

	pools := []struct {
		poolTypeId   uint32
		pairs        []string
		depositCoinA sdktypes.Int
		depositCoinB sdktypes.Int
	}{
		{
			uint32(1),
			[]string{
				"uakt",
				"uatom",
				"ubtsg",
				"udvpn",
				"ugcyb",
				"uiris",
				"uluna",
				"ungm",
				"uxprt",
				"uxrn",
				"xrun",
			},
			sdktypes.NewInt(1_000_000),
			sdktypes.NewInt(1_000_000),
		},
	}

	for _, p := range pools {
		totalSize := 0
		for i := len(p.pairs) - 1; i > 0; i-- {
			totalSize = totalSize + i
		}

		var msgs []sdktypes.Msg
		count := 0

		for i := 0; i < len(p.pairs)-1; i++ {
			for j := i + 1; j < len(p.pairs); j++ {
				count = count + 1
				log.Debug().Msgf("creating a pair of %s/%s, out of (%d/%d)", p.pairs[i], p.pairs[j], count, totalSize)

				depositCoins := sdktypes.NewCoins(
					sdktypes.NewCoin(p.pairs[i], p.depositCoinA),
					sdktypes.NewCoin(p.pairs[j], p.depositCoinB),
				)

				msg, err := transaction.MsgCreatePool(accAddr, p.poolTypeId, depositCoins)
				if err != nil {
					return fmt.Errorf("failed to create msg: %s", err)
				}
				msgs = append(msgs, msg)
			}
		}

		tx := transaction.NewTransaction(client, chainID, gasLimit, fees, memo)

		resp, err := tx.SignAndBroadcast(ctx, accAddr, privKey, msgs...)
		if err != nil {
			return fmt.Errorf("failed to sign and broadcast: %s", err)
		}

		if resp.TxResponse.Code != 0 {
			log.Debug().Msg("broadcasting transaction failed")
		}
		log.Debug().Msgf("reference: http://localhost:1317/cosmos/tx/v1beta1/txs/%s", resp.TxResponse.TxHash)
	}

	return nil
}

func bob() {

}
