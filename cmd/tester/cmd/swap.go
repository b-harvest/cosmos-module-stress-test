package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/b-harvest/cosmos-module-stress-test/client"
	"github.com/b-harvest/cosmos-module-stress-test/config"
	"github.com/b-harvest/cosmos-module-stress-test/tx"
	"github.com/b-harvest/cosmos-module-stress-test/wallet"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func SwapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "swap [pool-id] [offer-coin] [demand-coin-denom] [round] [tx-num] [msg-num]",
		Short:   "swap offer coin with demand coin.",
		Aliases: []string{"s"},
		Args:    cobra.ExactArgs(6),
		Long: `Swap offer coin with demand coin from the liquidity pool in round times with a number of tx and msg messages.

Example: $ tester s 1 5000000ubtsg uatom 5 5 2

round: how many rounds to run
tx-num: how many transactions to be included in a block
msg-num: how many transaction messages to be included in a transaction
`,
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

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("pool-id %s not a valid uint, input a valid unsigned 32-bit integer for pool-id", args[0])
			}

			offerCoin, err := sdktypes.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			err = offerCoin.Validate()
			if err != nil {
				return err
			}

			err = sdktypes.ValidateDenom(args[2])
			if err != nil {
				return err
			}

			round, err := strconv.Atoi(args[3])
			if err != nil {
				return fmt.Errorf("round must be integer: %s", args[3])
			}

			txNum, err := strconv.Atoi(args[4])
			if err != nil {
				return fmt.Errorf("tx-num must be integer: %s", args[4])
			}

			msgNum, err := strconv.Atoi(args[5])
			if err != nil {
				return fmt.Errorf("msg-num must be integer: %s", args[5])
			}

			chainID, err := client.RPC.GetNetworkChainID(ctx)
			if err != nil {
				return err
			}

			accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(cfg.Custom.Mnemonics[0], "")
			if err != nil {
				return err
			}

			gasLimit := uint64(cfg.Custom.GasLimit)
			fees := sdktypes.NewCoins(sdktypes.NewCoin(cfg.Custom.FeeDenom, sdktypes.NewInt(cfg.Custom.FeeAmount)))
			memo := cfg.Custom.Memo

			tx := tx.NewTransaction(client, chainID, gasLimit, fees, memo)

			for i := 0; i < round; i++ {
				var txBytes [][]byte

				account, err := client.GRPC.GetBaseAccountInfo(ctx, accAddr)
				if err != nil {
					return fmt.Errorf("failed to get account information: %s", err)
				}

				accSeq := account.GetSequence()
				accNum := account.GetAccountNumber()

				msgs, err := tx.CreateSwapBot(ctx, accAddr, poolId, offerCoin, args[2], msgNum)
				if err != nil {
					return fmt.Errorf("failed to create msg: %s", err)
				}

				for i := 0; i < txNum; i++ {
					txByte, err := tx.Sign(ctx, accSeq, accNum, privKey, msgs...)
					if err != nil {
						return fmt.Errorf("failed to sign and broadcast: %s", err)
					}

					accSeq = accSeq + 1

					txBytes = append(txBytes, txByte)
				}

				log.Info().Msgf("round:%d; txNum:%d; msgNum: %d; accAddr:%s", i+1, txNum, msgNum, accAddr)

				for _, txByte := range txBytes {
					resp, err := client.GRPC.BroadcastTx(ctx, txByte)
					if err != nil {
						return fmt.Errorf("failed to broadcast transaction: %s", err)
					}

					log.Info().Msgf("%s/cosmos/tx/v1beta1/txs/%s", cfg.LCD.Address, resp.TxResponse.TxHash)
				}
			}

			return nil
		},
	}
	return cmd
}
