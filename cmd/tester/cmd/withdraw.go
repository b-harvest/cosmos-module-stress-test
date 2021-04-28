package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/b-harvest/liquidity-stress-test/client"
	"github.com/b-harvest/liquidity-stress-test/config"
	"github.com/b-harvest/liquidity-stress-test/tx"
	"github.com/b-harvest/liquidity-stress-test/wallet"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

func WithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw [round]",
		Short:   "withdraw some coins in round times with mutiple messages in all existing pools.",
		Aliases: []string{"w"},
		Args:    cobra.ExactArgs(1),
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

			accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(cfg.Accounts.Withdraw, "")
			if err != nil {
				return fmt.Errorf("failed to retrieve account from mnemonic: %s", err)
			}

			pools, err := client.GRPC.GetAllPools(ctx)
			if err != nil {
				return fmt.Errorf("failed to get all liquidity pools: %s", err)
			}

			var msgs []sdktypes.Msg

			for _, pool := range pools {
				poolCoin := sdktypes.NewCoin(pool.PoolCoinDenom, sdktypes.NewInt(cfg.Amounts.Withdraw))

				msg, err := tx.MsgWithdraw(accAddr, pool.GetPoolId(), poolCoin)
				if err != nil {
					return fmt.Errorf("failed to create msg: %s", err)
				}
				msgs = append(msgs, msg)
			}

			round, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("round must be integer: %s", args[0])
			}

			account, err := client.GRPC.GetBaseAccountInfo(ctx, accAddr)
			if err != nil {
				return fmt.Errorf("failed to get account information: %s", err)
			}

			accSeq := account.GetSequence()
			accNum := account.GetAccountNumber()

			gasLimit := uint64(cfg.Tx.GasLimit)
			fees := sdktypes.NewCoins(sdktypes.NewCoin(cfg.Tx.FeeDenom, sdktypes.NewInt(cfg.Tx.FeeAmount)))
			memo := cfg.Tx.Memo

			tx := tx.NewTransaction(client, chainID, gasLimit, fees, memo)

			for i := 0; i < round; i++ {
				resp, err := tx.SignAndBroadcast(ctx, accSeq, accNum, privKey, msgs...)
				if err != nil {
					return fmt.Errorf("failed to sign and broadcast: %s", err)
				}

				accSeq = accSeq + 1

				log.Debug().
					Str("account", accAddr).
					Uint64("accSeq", accSeq).
					Int("round", i+1).
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
