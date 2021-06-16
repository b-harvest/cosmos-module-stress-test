package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/b-harvest/cosmos-module-stress-test/client"
	"github.com/b-harvest/cosmos-module-stress-test/config"
	"github.com/b-harvest/cosmos-module-stress-test/tx"
	"github.com/b-harvest/cosmos-module-stress-test/wallet"
)

func StressTestCmd() *cobra.Command {
	var (
		heightSpan     int64
		numTxsPerBlock int
		numMsgsPerTx   int
	)
	cmd := &cobra.Command{
		Use:   "stress-test [pool-id] [offer-coin] [starting-height]",
		Short: "run stress test",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			ctx := context.Background()

			err := SetLogger(logLevel)
			if err != nil {
				return fmt.Errorf("set logger: %w", err)
			}

			cfg, err := config.Read(config.DefaultConfigPath)
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}

			client, err := client.NewClient(cfg.RPC.Address, cfg.GRPC.Address)
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}
			defer client.Stop() // nolint: errcheck

			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pool id: %w", err)
			}

			offerCoin, err := sdktypes.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid offer coin: %w", err)
			}

			if err := offerCoin.Validate(); err != nil {
				return fmt.Errorf("invalid offer coin: %w", err)
			}

			startingHeight, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid starting height: %w", err)
			}

			pool, err := client.GRPC.GetPool(ctx, poolID)
			if err != nil {
				return fmt.Errorf("get pool: %w", err)
			}

			var demandCoinDenom string
			if pool.ReserveCoinDenoms[0] == offerCoin.Denom {
				demandCoinDenom = pool.ReserveCoinDenoms[1]
			} else {
				demandCoinDenom = pool.ReserveCoinDenoms[0]
			}

			chainID, err := client.RPC.GetNetworkChainID(ctx)
			if err != nil {
				return err
			}

			accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(cfg.Custom.Mnemonic, "")
			if err != nil {
				return err
			}

			gasLimit := uint64(cfg.Custom.GasLimit)
			fees := sdktypes.NewCoins(sdktypes.NewCoin(cfg.Custom.FeeDenom, sdktypes.NewInt(cfg.Custom.FeeAmount)))
			memo := cfg.Custom.Memo

			tx := tx.NewTransaction(client, chainID, gasLimit, fees, memo)

			log.Info().Msg("waiting until the block height reaches the starting height")
			for {
				st, err := client.RPC.Status(ctx)
				if err != nil {
					return fmt.Errorf("get status: %w", err)
				}
				if st.SyncInfo.LatestBlockHeight == startingHeight-1 {
					break
				} else if st.SyncInfo.LatestBlockHeight >= startingHeight {
					return fmt.Errorf("the block height has already past the starting height")
				}
			}
			log.Info().Msg("reached the starting height")

			for i := int64(0); i < heightSpan; i++ {
				nextHeight := startingHeight + i
				var txBytes [][]byte

				time.Sleep(time.Second) // Wait for the account sequence to increase.

				account, err := client.GRPC.GetBaseAccountInfo(ctx, accAddr)
				if err != nil {
					return fmt.Errorf("failed to get account information: %s", err)
				}

				accSeq := account.GetSequence()
				log.Info().Msgf("account sequence number: %d", accSeq)
				accNum := account.GetAccountNumber()

				msgs, err := tx.CreateSwapBot(ctx, accAddr, poolID, offerCoin, demandCoinDenom, numMsgsPerTx)
				if err != nil {
					return fmt.Errorf("failed to create msg: %s", err)
				}

				for i := 0; i < numTxsPerBlock; i++ {
					txByte, err := tx.Sign(ctx, accSeq, accNum, privKey, msgs...)
					if err != nil {
						return fmt.Errorf("failed to sign and broadcast: %s", err)
					}
					accSeq++
					txBytes = append(txBytes, txByte)
				}

				pendingTxHashes := make(map[string]struct{})

				for _, txByte := range txBytes {
					resp, err := client.GRPC.BroadcastTx(ctx, txByte)
					if err != nil {
						return fmt.Errorf("failed to broadcast transaction: %s", err)
					}
					pendingTxHashes[resp.TxResponse.TxHash] = struct{}{}
					//fmt.Print(".")
				}
				//fmt.Println()

				// Wait for the block to be committed.
				for {
					st, err := client.GetRPCClient().Status(ctx)
					if err != nil {
						return fmt.Errorf("get status: %w", err)
					}
					if st.SyncInfo.LatestBlockHeight == nextHeight {
						break
					} else if st.SyncInfo.LatestBlockHeight > nextHeight {
						return fmt.Errorf("block has past")
					}
					time.Sleep(100 * time.Millisecond)
				}

				r, err := client.GetRPCClient().Block(ctx, &nextHeight)
				if err != nil {
					return err
				}
				for _, tx := range r.Block.Txs {
					delete(pendingTxHashes, strings.ToUpper(hex.EncodeToString(tx.Hash())))
				}
				log.Info().Msgf("height: %d, block time: %s, number of txs: %d, number of missing txs: %d",
					nextHeight, r.Block.Time.Format(time.RFC3339Nano), len(r.Block.Txs), len(pendingTxHashes))

				time.Sleep(time.Second)
			}

			return nil
		},
	}
	cmd.Flags().Int64VarP(&heightSpan, "height-span", "s", 10, "how many blocks will the test run for")
	cmd.Flags().IntVarP(&numTxsPerBlock, "num-txs", "t", 1, "number of transactions per block")
	cmd.Flags().IntVarP(&numMsgsPerTx, "num-msgs", "m", 1, "number of messages per transaction")
	return cmd
}
