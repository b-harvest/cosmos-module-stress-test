package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	tenderminttypes "github.com/tendermint/tendermint/types"

	"github.com/b-harvest/cosmos-module-stress-test/client"
	"github.com/b-harvest/cosmos-module-stress-test/config"
	"github.com/b-harvest/cosmos-module-stress-test/tx"
	"github.com/b-harvest/cosmos-module-stress-test/wallet"
)

type Tx struct {
	Hash            string
	BroadcastHeight int64
	CommittedHeight int64
}

type Txs map[string]*Tx

type TxTracker struct {
	txsByHeight        map[int64]Txs
	pendingTxsByHeight map[int64]Txs
	txs                Txs
}

func NewTxTracker() *TxTracker {
	return &TxTracker{
		txsByHeight:        make(map[int64]Txs),
		pendingTxsByHeight: make(map[int64]Txs),
		txs:                make(Txs),
	}
}

func (tr *TxTracker) TxBroadcast(hash string, height int64) {
	tx := &Tx{
		Hash:            hash,
		BroadcastHeight: height,
	}
	if _, ok := tr.txsByHeight[height]; !ok {
		tr.txsByHeight[height] = make(Txs)
	}
	tr.txsByHeight[height][hash] = tx
	if _, ok := tr.pendingTxsByHeight[height]; !ok {
		tr.pendingTxsByHeight[height] = make(Txs)
	}
	tr.pendingTxsByHeight[height][hash] = tx
	tr.txs[hash] = tx
}

func (tr *TxTracker) TxsCommitted(hashes []string, height int64) (finishedHeights []int64) {
	for _, hash := range hashes {
		tx, ok := tr.txs[hash]
		if !ok {
			continue
		}
		tx.CommittedHeight = height
		delete(tr.pendingTxsByHeight[tx.BroadcastHeight], hash)
		if len(tr.pendingTxsByHeight[tx.BroadcastHeight]) == 0 {
			finishedHeights = append(finishedHeights, tx.BroadcastHeight)
		}
	}
	return
}

func (tr *TxTracker) AllDelays(currentHeight int64) (delays []int64) {
	for _, tx := range tr.txs {
		var delay int64
		if tx.CommittedHeight != 0 {
			delay = tx.CommittedHeight - tx.BroadcastHeight
		} else {
			delay = currentHeight - tx.BroadcastHeight
		}
		delays = append(delays, delay)
	}
	return
}

func (tr *TxTracker) Delays(currentHeight int64) (delays []int64) {
	for _, tx := range tr.txs {
		var delay int64
		if tx.CommittedHeight != 0 {
			delay = tx.CommittedHeight - tx.BroadcastHeight
		} else {
			delay = currentHeight - tx.BroadcastHeight
		}
		if delay > 0 {
			delays = append(delays, currentHeight-tx.BroadcastHeight)
		}
	}
	return
}

func (tr *TxTracker) NumMissedTxs(currentHeight int64) int {
	n := 0
	for _, tx := range tr.txs {
		if tx.CommittedHeight != 0 {
			continue
		}
		if currentHeight-tx.BroadcastHeight > 5 {
			n++
		}
	}
	return n
}

func TxHashes(txs tenderminttypes.Txs) []string {
	var hashes []string
	for _, tx := range txs {
		hashes = append(hashes, strings.ToUpper(hex.EncodeToString(tx.Hash())))
	}
	return hashes
}

func AvgDelays(delays []int64) float64 {
	if len(delays) == 0 {
		return 0
	}
	sum := 0.0
	for _, delay := range delays {
		sum += float64(delay)
	}
	return sum / float64(len(delays))
}

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

			offerCoin, err := sdk.ParseCoinNormalized(args[1])
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
			fees := sdk.NewCoins(sdk.NewCoin(cfg.Custom.FeeDenom, sdk.NewInt(cfg.Custom.FeeAmount)))
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
				time.Sleep(100 * time.Millisecond)
			}
			log.Info().Msg("reached the starting height")

			acc, err := client.GRPC.GetBaseAccountInfo(ctx, accAddr)
			if err != nil {
				return fmt.Errorf("get base account info: %w", err)
			}
			seq := acc.GetSequence()

			tr := NewTxTracker()
			prevSeq := uint64(0)

			broadcastTxHashes := make(map[string]struct{})

			for i := int64(0); i < heightSpan; i++ {
				fmt.Println(strings.Repeat("=", 70))

				nextHeight := startingHeight + i
				var txBytes [][]byte

				started := time.Now()
				var account types.BaseAccount
				for {
					var err error
					account, err = client.GRPC.GetBaseAccountInfo(ctx, accAddr)
					if err != nil {
						return fmt.Errorf("failed to get account information: %s", err)
					}
					if account.GetSequence() > prevSeq {
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
				log.Debug().Msgf("took %s waiting for the account info update", time.Since(started))

				accSeq := account.GetSequence()
				log.Info().Msgf("account sequence: %d", accSeq)
				prevSeq = accSeq
				accNum := account.GetAccountNumber()

				msgs, err := tx.CreateSwapBot(ctx, accAddr, poolID, offerCoin, demandCoinDenom, numMsgsPerTx)
				if err != nil {
					return fmt.Errorf("failed to create msg: %s", err)
				}

				started = time.Now()
				for i := 0; i < numTxsPerBlock; i++ {
					txByte, err := tx.Sign(ctx, seq, accNum, privKey, msgs...)
					if err != nil {
						return fmt.Errorf("failed to sign and broadcast: %s", err)
					}
					seq++
					txBytes = append(txBytes, txByte)
				}
				log.Debug().Msgf("took %s signing txs", time.Since(started))

				started = time.Now()
				numFailedTxs := 0
				for _, txByte := range txBytes {
					resp, err := client.GRPC.BroadcastTx(ctx, txByte)
					if err != nil {
						return fmt.Errorf("failed to broadcast transaction: %s", err)
					}
					if _, ok := broadcastTxHashes[resp.TxResponse.TxHash]; ok {
						panic("duplicate tx")
					}
					broadcastTxHashes[resp.TxResponse.TxHash] = struct{}{}
					if resp.TxResponse.Code != 0 {
						numFailedTxs++
						if resp.TxResponse.Code != 19 {
							panic(fmt.Sprintf("%#v\n", resp.TxResponse))
						}
					}
					tr.TxBroadcast(resp.TxResponse.TxHash, nextHeight)
					//fmt.Print(".")
				}
				//fmt.Println()
				log.Debug().Msgf("took %s broadcasting txs", time.Since(started))
				if numFailedTxs > 0 {
					log.Info().Msgf("number of failed txs: %d", numFailedTxs)
				}

				// Wait for the block to be committed.
				started = time.Now()
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
				log.Debug().Msgf("took %s waiting for the next block", time.Since(started))

				r, err := client.GetRPCClient().Block(ctx, &nextHeight)
				if err != nil {
					return err
				}
				log.Info().Msgf("height: %d, block time: %s, number of txs: %d",
					nextHeight, r.Block.Time.Format(time.RFC3339Nano), len(r.Block.Txs))

				finishedHeights := tr.TxsCommitted(TxHashes(r.Block.Txs), nextHeight)
				if len(finishedHeights) > 0 {
					log.Info().Msgf("finished heights: %v", finishedHeights)
				}

				log.Info().Msgf("all avg: %f, avg: %f, missing: %d", AvgDelays(tr.AllDelays(nextHeight)), AvgDelays(tr.Delays(nextHeight)), tr.NumMissedTxs(nextHeight))

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
