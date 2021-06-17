package cmd

import (
	"context"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (tr *TxTracker) Delays(height int64) (delays []int64) {
	for _, tx := range tr.txsByHeight[height] {
		if tx.CommittedHeight == 0 {
			panic("delays must be calculated on the finished block")
		}
		delay := tx.CommittedHeight - tx.BroadcastHeight
		delays = append(delays, delay)
	}
	return
}

func (tr *TxTracker) NumMissingTxs(height int64) int {
	n := 0
	for _, tx := range tr.txsByHeight[height] {
		if tx.CommittedHeight != tx.BroadcastHeight {
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

type AccountDispenser struct {
	c         *client.Client
	mnemonics []string
	i         int
	addr      string
	privKey   *secp256k1.PrivKey
	accSeq    uint64
	accNum    uint64
}

func NewAccountDispenser(c *client.Client, mnemonics []string) *AccountDispenser {
	return &AccountDispenser{
		c:         c,
		mnemonics: mnemonics,
	}
}

func (d *AccountDispenser) Next() error {
	mnemonic := d.mnemonics[d.i]
	addr, privKey, err := wallet.RecoverAccountFromMnemonic(mnemonic, "")
	if err != nil {
		return err
	}
	d.addr = addr
	d.privKey = privKey
	acc, err := d.c.GRPC.GetBaseAccountInfo(context.Background(), addr)
	if err != nil {
		return fmt.Errorf("get base account info: %w", err)
	}
	d.accSeq = acc.GetSequence()
	d.accNum = acc.GetAccountNumber()
	d.i++
	if d.i >= len(d.mnemonics) {
		d.i = 0
	}
	return nil
}

func (d *AccountDispenser) Addr() string {
	return d.addr
}

func (d *AccountDispenser) PrivKey() *secp256k1.PrivKey {
	return d.privKey
}

func (d *AccountDispenser) AccSeq() uint64 {
	return d.accSeq
}

func (d *AccountDispenser) AccNum() uint64 {
	return d.accNum
}

func (d *AccountDispenser) IncAccSeq() uint64 {
	r := d.accSeq
	d.accSeq++
	return r
}

func StressTestCmd() *cobra.Command {
	var (
		heightSpan     int64
		numTxsPerBlock int
		numMsgsPerTx   int
	)
	cmd := &cobra.Command{
		Use:   "stress-test [pool-id] [offer-coin]",
		Short: "run stress test",
		Args:  cobra.ExactArgs(2),
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

			d := NewAccountDispenser(client, cfg.Custom.Mnemonics)

			st, err := client.RPC.Status(ctx)
			if err != nil {
				return fmt.Errorf("get status: %w", err)
			}
			currentHeight := st.SyncInfo.LatestBlockHeight
			log.Info().Msgf("current height: %d, waiting for the next block", currentHeight)

			blockTimes := make(map[int64]time.Time)

			var startingHeight int64
			for {
				st, err := client.RPC.Status(ctx)
				if err != nil {
					return fmt.Errorf("get status: %w", err)
				}
				blockTimes[st.SyncInfo.LatestBlockHeight] = st.SyncInfo.LatestBlockTime
				if st.SyncInfo.LatestBlockHeight > currentHeight {
					startingHeight = st.SyncInfo.LatestBlockHeight + 1
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			log.Info().Msgf("starting at %d", startingHeight)

			if err := d.Next(); err != nil {
				return fmt.Errorf("next account: %w", err)
			}
			log.Info().Msgf("starting sequence: %d", d.AccSeq())

			gasLimit := uint64(cfg.Custom.GasLimit)
			fees := sdk.NewCoins(sdk.NewCoin(cfg.Custom.FeeDenom, sdk.NewInt(cfg.Custom.FeeAmount)))
			memo := cfg.Custom.Memo
			tx := tx.NewTransaction(client, chainID, gasLimit, fees, memo)

			tr := NewTxTracker()

			f, err := os.OpenFile("result.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}
			defer f.Close()

			w := csv.NewWriter(f)
			defer w.Flush()

			for i := int64(0); i < heightSpan; i++ {
				fmt.Println(strings.Repeat("-", 80))

				nextHeight := startingHeight + i

				msgs, err := tx.CreateSwapBot(ctx, d.Addr(), poolID, offerCoin, demandCoinDenom, numMsgsPerTx)
				if err != nil {
					return fmt.Errorf("failed to create msg: %s", err)
				}

				started := time.Now()
				for i := 0; i < numTxsPerBlock; i++ {
					accSeq := d.IncAccSeq()
					txByte, err := tx.Sign(ctx, accSeq, d.AccNum(), d.PrivKey(), msgs...)
					if err != nil {
						return fmt.Errorf("sign tx: %w", err)
					}
					resp, err := client.GRPC.BroadcastTx(ctx, txByte)
					if err != nil {
						return fmt.Errorf("broadcast tx: %w", err)
					}
					if resp.TxResponse.Code != 0 {
						if resp.TxResponse.Code == 0x13 || resp.TxResponse.Code == 0x20 {
							log.Warn().Msgf("received %#v, using the next account", resp.TxResponse)
							if err := d.Next(); err != nil {
								return fmt.Errorf("next account: %w", err)
							}
							log.Warn().Msgf("next account address: %s", d.Addr())
						} else {
							panic(fmt.Sprintf("%#v\n", resp.TxResponse))
						}
					} else {
						tr.TxBroadcast(resp.TxResponse.TxHash, nextHeight)
					}
				}
				log.Debug().Msgf("took %s broadcasting txs", time.Since(started))

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
				blockTimes[nextHeight] = r.Block.Time
				log.Info().Msgf("height: %d, block time: %s, number of txs: %d",
					nextHeight, r.Block.Time.Format(time.RFC3339Nano), len(r.Block.Txs))

				finishedHeights := tr.TxsCommitted(TxHashes(r.Block.Txs), nextHeight)
				log.Info().Msgf("finished heights: %v", finishedHeights)

				for _, height := range finishedHeights {
					blockTime := blockTimes[height]
					blockDuration := blockTimes[height].Sub(blockTimes[height-1])
					numMissingTxs := tr.NumMissingTxs(height)
					avgDelay := AvgDelays(tr.Delays(height))

					log.Info().Msgf("> height: %d, block time: %s, block duration: %s, num missing txs: %d, avg delay: %f",
						height, blockTime.Format(time.RFC3339Nano), blockDuration, numMissingTxs, avgDelay)
					if err := w.Write([]string{
						strconv.FormatInt(height, 10),
						blockTime.Format(time.RFC3339Nano),
						blockDuration.String(),
						strconv.Itoa(numMissingTxs),
						strconv.FormatFloat(avgDelay, 'f', -1, 64),
					}); err != nil {
						return fmt.Errorf("write record")
					}
					w.Flush()
				}
			}

			return nil
		},
	}
	cmd.Flags().Int64VarP(&heightSpan, "height-span", "s", 10, "how many blocks will the test run for")
	cmd.Flags().IntVarP(&numTxsPerBlock, "num-txs", "t", 1, "number of transactions per block")
	cmd.Flags().IntVarP(&numMsgsPerTx, "num-msgs", "m", 1, "number of messages per transaction")
	return cmd
}
