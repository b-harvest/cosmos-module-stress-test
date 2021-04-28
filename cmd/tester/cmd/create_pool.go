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

	liqtypes "github.com/tendermint/liquidity/x/liquidity/types"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

// CreateAllPoolsCmd creates liquidity pools of every pair of coins exist in the network.
// This command is useful for stress testing to bootstrap test pools as soon as new network is spun up.
// Since the Gravity DEX testnet will have 11 coin types available, this will create a total number of 55 pairs of pools.
func CreateAllPoolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-all-pools",
		Short:   "create liquidity pools of every pair of coins exist in the network.",
		Aliases: []string{"create-all", "ca"},
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

			cfg, err := config.Read(config.DefaultConfigPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %s", err)
			}

			client, err := client.NewClient(cfg.RPC.Address, cfg.GRPC.Address)
			if err != nil {
				return fmt.Errorf("failed to connect clients: %s", err)
			}

			defer client.Stop() // nolint: errcheck

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			chainID, err := client.RPC.GetNetworkChainID(ctx)
			if err != nil {
				return fmt.Errorf("failed to get chain id: %s", err)
			}

			accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(cfg.Custom.Mnemonic, "")
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
					liqtypes.DefaultPoolTypeId,
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
					sdktypes.NewInt(50_000_000),
					sdktypes.NewInt(50_000_000),
				},
			}

			for _, p := range pools {
				totalNum := 0
				count := 0

				for i := len(p.pairs) - 1; i > 0; i-- {
					totalNum = totalNum + i
				}

				var msgs []sdktypes.Msg

				// find all pairs of coins. {coinA, coinB} and {coinB coinA} are excluded.
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
