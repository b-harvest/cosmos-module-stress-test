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

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

var (
	configPath = "./config.toml"

	// cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v
	mnemonic1 = "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"

	feeDenom = "stake"
)

func init() {
	// configure log level and use pretty logging
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	log.Info().Msg("> Liquidity Stress Testing Begins...")

	cfg, err := config.Read(configPath)
	if err != nil {
		log.Panic().Err(err).Msg("failed to read config file")
	}

	ctx := context.Background()
	client := client.NewClient(cfg.RPC.Address, cfg.GRPC.Address)

	chainID, err := client.RPC.GetNetworkChainID(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get chain id")
		return
	}

	tx := transaction.NewTransaction(client, chainID, mnemonic1, feeDenom)

	// MsgCreatePool
	depositCoinA := sdktypes.NewCoin("uakt", sdktypes.NewInt(1000000))
	depositCoinB := sdktypes.NewCoin("uatom", sdktypes.NewInt(1000000))
	resp, err := tx.SignAndBroadcastMsgCreatePool(ctx, depositCoinA, depositCoinB)
	if err != nil {
		log.Error().Err(err).Msg("failed to get chain id")
		return
	}

	fmt.Println("resp: ", resp)
}
