package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nodebreaker0-0/cosmos-module-stress-test/client"
	"github.com/nodebreaker0-0/cosmos-module-stress-test/config"
	"github.com/nodebreaker0-0/cosmos-module-stress-test/tx"
	"github.com/nodebreaker0-0/cosmos-module-stress-test/wallet"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

const (
	flagPacketTimeoutHeight    = "packet-timeout-height"
	flagPacketTimeoutTimestamp = "packet-timeout-timestamp"
	flagAbsoluteTimeouts       = "absolute-timeouts"
)

func IbctransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transfer [src-port] [src-channel] [receiver] [amount] [round] [tx-num] [msg-num]",
		Short:   "Transfer a fungible token through IBC",
		Aliases: []string{"t"},
		Args:    cobra.ExactArgs(7),
		Long: `Transfer a fungible token through IBC.

Example: $tester t transfer channel-0 cosmos1pacc0fr45hggcn8jrfhgnqf8vgyqna7r5sftql 10uatom 10 1 1

[round]: how many rounds to run
[tx-num]: how many transactions to be included in one round
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

			if err != nil {
				return err
			}
			client, err := client.NewClient(cfg.RPC.Address, cfg.GRPC.Address)
			if err != nil {
				return fmt.Errorf("failed to connect clients: %s", err)
			}

			defer client.Stop() // nolint: errcheck
			ibcclientCtx := client.GetCLIContext()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			chainID, err := client.RPC.GetNetworkChainID(ctx)
			if err != nil {
				return fmt.Errorf("failed to get chain id: %s", err)
			}

			srcPort := args[0]
			srcChannel := args[1]
			receiver := args[2]

			coin, err := sdktypes.ParseCoinNormalized(args[3])
			if err != nil {
				return err
			}

			if !strings.HasPrefix(coin.Denom, "ibc/") {
				denomTrace := ibctypes.ParseDenomTrace(coin.Denom)
				coin.Denom = denomTrace.IBCDenom()
			}

			round, err := strconv.Atoi(args[4])
			if err != nil {
				return fmt.Errorf("round must be integer: %s", args[0])
			}

			txNum, err := strconv.Atoi(args[5])
			if err != nil {
				return fmt.Errorf("txNum must be integer: %s", args[0])
			}

			msgNum, err := strconv.Atoi(args[6])
			if err != nil {
				return fmt.Errorf("txNum must be integer: %s", args[0])
			}

			accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(cfg.Custom.Mnemonic, "")
			if err != nil {
				return fmt.Errorf("failed to retrieve account from mnemonic: %s", err)
			}

			gasLimit := uint64(cfg.Custom.GasLimit)
			fees := sdktypes.NewCoins(sdktypes.NewCoin(cfg.Custom.FeeDenom, sdktypes.NewInt(cfg.Custom.FeeAmount)))
			memo := cfg.Custom.Memo

			tx := tx.IbcNewtransaction(client, chainID, gasLimit, fees, memo)

			for i := 0; i < round; i++ {
				var txBytes [][]byte

				account, err := client.GRPC.GetBaseAccountInfo(ctx, accAddr)
				if err != nil {
					return fmt.Errorf("failed to get account information: %s", err)
				}

				accSeq := account.GetSequence()
				accNum := account.GetAccountNumber()

				msgs, err := tx.CreateTransferBot(cmd, ibcclientCtx, srcPort, srcChannel, coin, accAddr, receiver, msgNum)

				if err != nil {
					return fmt.Errorf("failed to create msg: %s", err)
				}

				for _, msg := range msgs {
					fmt.Println("msg: ", msg)
				}

				for i := 0; i < txNum; i++ {
					txByte, err := tx.IbcSign(ctx, accSeq, accNum, privKey, msgs...)
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
	cmd.Flags().String(flagPacketTimeoutHeight, ibctypes.DefaultRelativePacketTimeoutHeight, "Packet timeout block height. The timeout is disabled when set to 0-0.")
	cmd.Flags().Uint64(flagPacketTimeoutTimestamp, ibctypes.DefaultRelativePacketTimeoutTimestamp, "Packet timeout timestamp in nanoseconds. Default is 10 minutes. The timeout is disabled when set to 0.")
	cmd.Flags().Bool(flagAbsoluteTimeouts, false, "Timeout flags are used as absolute timeouts.")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
