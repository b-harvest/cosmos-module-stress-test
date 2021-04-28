package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

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

func SwapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "swap [round]",
		Short:   "swap some coins in round times with mutiple messages in all existing pools.",
		Aliases: []string{"s"},
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

			chainID, err := client.RPC.GetNetworkChainID(context.Background())
			if err != nil {
				return fmt.Errorf("failed to get chain id: %s", err)
			}

			accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(cfg.Accounts.Swap, "")
			if err != nil {
				return fmt.Errorf("failed to retrieve account from mnemonic: %s", err)
			}

			pools, err := client.GRPC.GetAllPools(context.Background())
			if err != nil {
				return fmt.Errorf("failed to get all liquidity pools: %s", err)
			}

			var msgs []sdktypes.Msg

			for _, pool := range pools {
				swapTypeId := liqtypes.DefaultSwapTypeId
				offerCoin := sdktypes.NewCoin(pool.ReserveCoinDenoms[0], sdktypes.NewInt(cfg.Amounts.Swap))
				demandCoinDenom := pool.ReserveCoinDenoms[1]
				orderPrice := sdktypes.NewDecWithPrec(19, 3)
				swapFeeRate := sdktypes.NewDecWithPrec(3, 3)

				msg, err := tx.MsgSwap(accAddr, pool.GetPoolId(), swapTypeId, offerCoin, demandCoinDenom, orderPrice, swapFeeRate)
				if err != nil {
					return fmt.Errorf("failed to create msg: %s", err)
				}
				msgs = append(msgs, msg)
			}

			round, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("round must be integer: %s", args[0])
			}

			account, err := client.GRPC.GetBaseAccountInfo(context.Background(), accAddr)
			if err != nil {
				return fmt.Errorf("failed to get account information: %s", err)
			}

			accSeq := account.GetSequence()
			accNum := account.GetAccountNumber()

			tx := tx.NewTransaction(client, chainID, tx.DefaultGasLimit, tx.DefaultFees, tx.DefaultMemo)

			for i := 0; i < round; i++ {
				resp, err := tx.SignAndBroadcast(accSeq, accNum, privKey, msgs...)
				if err != nil {
					return fmt.Errorf("failed to sign and broadcast: %s", err)
				}

				accSeq = accSeq + 1

				log.Debug().
					Str("account", accAddr).
					Uint64("account sequence", accSeq).
					Int("round", i+1).
					Str("number of messsages", fmt.Sprintf("%d", len(msgs))).
					Uint32("code", resp.TxResponse.Code).
					Int64("height", resp.TxResponse.Height).
					Str("hash", resp.TxResponse.TxHash).
					Msg("swap result")

				log.Info().Msgf("reference: http://localhost:1317/cosmos/tx/v1beta1/txs/%s", resp.TxResponse.TxHash)

				time.Sleep(5 * time.Second)
			}

			return nil
		},
	}
	return cmd
}
