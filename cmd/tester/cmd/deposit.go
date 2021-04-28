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

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

func DepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deposit [pool-id] [deposit-coins] [round] [msg-num]",
		Short:   "deposit coins to a liquidity pool in round times with a number of transaction messages",
		Aliases: []string{"d"},
		Args:    cobra.ExactArgs(4),
		Long: `Deposit coins a liquidity pool in round times.

Example: $tester d 1 100000000uatom,5000000000uusd 10 10
`,
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

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			chainID, err := client.RPC.GetNetworkChainID(ctx)
			if err != nil {
				return fmt.Errorf("failed to get chain id: %s", err)
			}

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("pool-id %s not a valid uint, input a valid unsigned 32-bit integer for pool-id", args[0])
			}

			depositCoins, err := sdktypes.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			err = depositCoins.Validate()
			if err != nil {
				return err
			}

			if depositCoins.Len() != 2 {
				return fmt.Errorf("the number of deposit coins must be two in the pool-type 1")
			}

			round, err := strconv.Atoi(args[2])
			if err != nil {
				return fmt.Errorf("round must be integer: %s", args[0])
			}

			msgNum, err := strconv.Atoi(args[3])
			if err != nil {
				return fmt.Errorf("txNum must be integer: %s", args[0])
			}

			for i := 0; i < round; i++ {
				accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(cfg.Custom.Mnemonic, "")
				if err != nil {
					return fmt.Errorf("failed to retrieve account from mnemonic: %s", err)
				}

				msg, err := tx.MsgDeposit(accAddr, poolId, depositCoins)
				if err != nil {
					return fmt.Errorf("failed to create msg: %s", err)
				}

				msgs := []sdktypes.Msg{msg}

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

				var txBytes [][]byte

				for i := 0; i < msgNum; i++ {
					txByte, err := tx.Sign(ctx, accSeq, accNum, privKey, msgs...)
					if err != nil {
						return fmt.Errorf("failed to sign and broadcast: %s", err)
					}

					accSeq = accSeq + 1

					txBytes = append(txBytes, txByte)
				}

				log.Info().Msgf("round:%d; msgNum:%d; accAddr:%s", i+1, msgNum, accAddr)

				for _, txByte := range txBytes {
					resp, err := client.GRPC.BroadcastTx(ctx, txByte)
					if err != nil {
						return fmt.Errorf("failed to broadcast transaction: %s", err)
					}

					log.Info().Msgf("%s/cosmos/tx/v1beta1/txs/%s", cfg.LCD.Address, resp.TxResponse.TxHash)
				}

				// wait for one block
				time.Sleep(1 * time.Second)
			}

			return nil
		},
	}
	return cmd
}
