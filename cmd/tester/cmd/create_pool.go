package cmd

import (
	"context"
	"fmt"

	"github.com/b-harvest/cosmos-module-stress-test/client"
	"github.com/b-harvest/cosmos-module-stress-test/config"
	"github.com/b-harvest/cosmos-module-stress-test/tx"
	"github.com/b-harvest/cosmos-module-stress-test/wallet"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	liqtypes "github.com/tendermint/liquidity/x/liquidity/types"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// The Gravity DEX testnet has 11 denom types available.
var denomPairs = []string{
	"uatom",
	"uiris",
	"ukava",
	"uluna",
	"uscrt",
}

// CreatePoolsCmd creates liquidity pools of every pair of coins exist in the network.
// This command is useful for stress testing to bootstrap test pools as soon as new network is spun up.
func CreatePoolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-pools",
		Short:   "create liquidity pools with the sample denom pairs.",
		Aliases: []string{"create", "c", "cp"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := SetLogger(logLevel)
			if err != nil {
				return err
			}

			cfg, err := config.Read(config.DefaultConfigPath)
			if err != nil {
				return err
			}

			client, err := client.NewClient(cfg.RPC.Address, cfg.GRPC.Address)
			if err != nil {
				return err
			}
			defer client.Stop() // nolint: errcheck

			chainID, err := client.RPC.GetNetworkChainID(ctx)
			if err != nil {
				return err
			}

			accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(cfg.Custom.Mnemonics[0], "")
			if err != nil {
				return err
			}

			pools := []struct {
				poolTypeId   uint32
				denomPairs   []string
				depositCoinA sdktypes.Int
				depositCoinB sdktypes.Int
			}{
				{
					liqtypes.DefaultPoolTypeId,
					denomPairs,
					sdktypes.NewInt(1_000_000_000),
					sdktypes.NewInt(1_000_000_000),
				},
			}

			for _, p := range pools {
				totalNum := 0
				count := 0

				for i := len(p.denomPairs) - 1; i > 0; i-- {
					totalNum = totalNum + i
				}

				var msgs []sdktypes.Msg

				// find all pairs of coins. {coinA, coinB} and {coinB coinA} are excluded.
				for i := 0; i < len(p.denomPairs)-1; i++ {
					for j := i + 1; j < len(p.denomPairs); j++ {
						count = count + 1
						log.Debug().Msgf("creating a pair of %s/%s, out of (%d/%d)", p.denomPairs[i], p.denomPairs[j], count, totalNum)

						depositCoins := sdktypes.NewCoins(
							sdktypes.NewCoin(p.denomPairs[i], p.depositCoinA),
							sdktypes.NewCoin(p.denomPairs[j], p.depositCoinB),
						)

						msg, err := tx.MsgCreatePool(accAddr, p.poolTypeId, depositCoins)
						if err != nil {
							return fmt.Errorf("failed to create msg: %s", err)
						}
						msgs = append(msgs, msg)
					}
				}

				account, err := client.GRPC.GetBaseAccountInfo(ctx, accAddr)
				if err != nil {
					return fmt.Errorf("failed to get account information: %s", err)
				}

				accSeq := account.GetSequence()
				accNum := account.GetAccountNumber()

				gasLimit := uint64(cfg.Custom.GasLimit)
				fees := sdktypes.NewCoins(sdktypes.NewCoin(cfg.Custom.FeeDenom, sdktypes.NewInt(cfg.Custom.FeeAmount)))
				memo := cfg.Custom.Memo

				tx := tx.NewTransaction(client, chainID, gasLimit, fees, memo)

				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				txBytes, err := tx.Sign(ctx, accSeq, accNum, privKey, msgs...)
				if err != nil {
					return fmt.Errorf("failed to sign and broadcast: %s", err)
				}

				resp, err := client.GRPC.BroadcastTx(ctx, txBytes)
				if err != nil {
					return fmt.Errorf("failed to broadcast transaction: %s", err)
				}

				log.Debug().
					Str("total messsages", fmt.Sprintf("%d", len(msgs))).
					Uint32("code", resp.TxResponse.Code).
					Int64("height", resp.TxResponse.Height).
					Str("hash", resp.TxResponse.TxHash).
					Msg("result")

				log.Info().Msgf("reference: %s/cosmos/tx/v1beta1/txs/%s", cfg.LCD.Address, resp.TxResponse.TxHash)
			}

			return nil
		},
	}
	return cmd
}
