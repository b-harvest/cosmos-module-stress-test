package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/b-harvest/liquidity-stress-test/client"
	"github.com/b-harvest/liquidity-stress-test/config"
	"github.com/b-harvest/liquidity-stress-test/tx"
	"github.com/b-harvest/liquidity-stress-test/wallet"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

// CreatePoolCmd creates liquidity pools of every pair of coins exist in the network.
// The Gravity DEX testnet will have 11 coin types available and
// this will create a total number of 55 pairs of liquidity pools.
func CreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-pool",
		Aliases: []string{"create", "c", "cp"},
		Short:   "create liquidity pools of every pair of coins exist in the network.",
		RunE: func(cmd *cobra.Command, args []string) error {
			logLvl, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				return err
			}

			zerolog.SetGlobalLevel(logLvl)

			switch logFormat {
			case logLevelJSON:
			case logLevelText:
				// human-readable pretty logging is the default logging format
				log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
			default:
				return fmt.Errorf("invalid logging format: %s", logFormat)
			}

			ctx := context.Background()

			cfg, err := config.Read(config.DefaultConfigPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %s", err)
			}

			client := client.NewClient(cfg.RPC.Address, cfg.GRPC.Address)

			chainID, err := client.RPC.GetNetworkChainID(ctx)
			if err != nil {
				return fmt.Errorf("failed to get chain id: %s", err)
			}

			accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(createPoolTester, "")
			if err != nil {
				return fmt.Errorf("failed to retrieve account and private key from mnemonic: %s", err)
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
					sdktypes.NewInt(50_000_000_000),
					sdktypes.NewInt(50_000_000_000),
				},
			}

			for _, p := range pools {
				totalNum := 0
				count := 0

				for i := len(p.pairs) - 1; i > 0; i-- {
					totalNum = totalNum + i
				}

				var msgs []sdktypes.Msg

				for i := 0; i < len(p.pairs)-1; i++ {
					for j := i + 1; j < len(p.pairs); j++ {
						count = count + 1
						log.Debug().Msgf("creating a pair of %s/%s, out of (%d/%d)", p.pairs[i], p.pairs[j], count, totalNum)

						depositCoins := sdktypes.NewCoins(
							sdktypes.NewCoin(p.pairs[i], p.depositCoinA),
							sdktypes.NewCoin(p.pairs[j], p.depositCoinB),
						)

						msg, err := tx.MsgCreatePool(accAddr, p.poolTypeId, depositCoins)
						if err != nil {
							return fmt.Errorf("failed to create msg: %s", err)
						}
						msgs = append(msgs, msg)
					}
				}

				tx := tx.NewTransaction(client, chainID, tx.DefaultGasLimit, tx.DefaultFees, tx.DefaultMemo)

				resp, err := tx.SignAndBroadcast(ctx, accAddr, privKey, msgs...)
				if err != nil {
					return fmt.Errorf("failed to sign and broadcast: %s", err)
				}

				log.Debug().
					Str("total number of sent messages", fmt.Sprintf("%d", len(msgs))).
					Uint32("code", resp.TxResponse.Code).
					Int64("height", resp.TxResponse.Height).
					Str("hash", resp.TxResponse.TxHash).
					Msg("create-pool tester result")

				log.Info().Msgf("reference: http://localhost:1317/cosmos/tx/v1beta1/txs/%s", resp.TxResponse.TxHash)
			}

			return nil
		},
	}
	return cmd
}
