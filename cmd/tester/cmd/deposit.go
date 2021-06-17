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

func DepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deposit [pool-id] [deposit-coins] [round] [tx-num]",
		Short:   "deposit coins to a liquidity pool in round times with a number of transaction messages",
		Aliases: []string{"d"},
		Args:    cobra.ExactArgs(4),
		Long: `Deposit coins to a liquidity pool in round times with a number of transaction messages.

Example: $ tester d 1 100000000uatom,5000000000uusd 10 10

[round]: how many rounds to run
[tx-num]: how many transactions to be included in one round
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
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
				return fmt.Errorf("round must be integer: %s", args[2])
			}

			txNum, err := strconv.Atoi(args[3])
			if err != nil {
				return fmt.Errorf("tx-num must be integer: %s", args[3])
			}

			chainID, err := client.RPC.GetNetworkChainID(ctx)
			if err != nil {
				return err
			}

			accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(cfg.Custom.Mnemonics[0], "")
			if err != nil {
				return err
			}

			msg, err := tx.MsgDeposit(accAddr, poolId, depositCoins)
			if err != nil {
				return fmt.Errorf("failed to create msg: %s", err)
			}

			msgs := []sdktypes.Msg{msg}

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

				for j := 0; j < txNum; j++ {
					txByte, err := tx.Sign(ctx, accSeq, accNum, privKey, msgs...)
					if err != nil {
						return fmt.Errorf("failed to sign and broadcast: %s", err)
					}

					accSeq = accSeq + 1

					txBytes = append(txBytes, txByte)
				}

				log.Info().Msgf("round:%d; txNum:%d; accAddr:%s", i+1, txNum, accAddr)

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
