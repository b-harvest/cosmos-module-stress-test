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

func DepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deposit",
		Aliases: []string{"d"},
		Short:   "deposit new coins to every existing pools.",
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

			accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(depositTester, "")
			if err != nil {
				return fmt.Errorf("failed to retrieve account from mnemonic: %s", err)
			}

			pools, err := client.GRPC.GetAllPools(ctx)
			if err != nil {
				return fmt.Errorf("failed to get all liquidity pools: %s", err)
			}

			var msgs []sdktypes.Msg

			for _, pool := range pools {
				depositCoins := sdktypes.NewCoins(
					sdktypes.NewCoin(pool.ReserveCoinDenoms[0], sdktypes.NewInt(1)),
					sdktypes.NewCoin(pool.ReserveCoinDenoms[1], sdktypes.NewInt(1)),
				)

				msg, err := tx.MsgDeposit(accAddr, pool.GetPoolId(), depositCoins)
				if err != nil {
					return fmt.Errorf("failed to create msg: %s", err)
				}
				msgs = append(msgs, msg)
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
				Msg("deposit tester result")

			log.Info().Msgf("reference: http://localhost:1317/cosmos/tx/v1beta1/txs/%s", resp.TxResponse.TxHash)

			return nil
		},
	}
	return cmd
}
