package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/b-harvest/cosmos-module-stress-test/client"
	"github.com/b-harvest/cosmos-module-stress-test/config"
	"github.com/b-harvest/cosmos-module-stress-test/tx"
	"github.com/b-harvest/cosmos-module-stress-test/wallet"
)

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

func (d *AccountDispenser) DecAccSeq() {
	d.accSeq--
}

type Scenario struct {
	Rounds         int
	NumTxsPerBlock int
}

var (
	scenarios = []Scenario{
		{100, 100},
		{100, 200},
		{100, 300},
		{100, 400},
		{100, 500},
	}
	//scenarios = []Scenario{
	//	{5, 10},
	//	{5, 50},
	//	{5, 100},
	//	{5, 200},
	//	{5, 300},
	//	{5, 400},
	//	{5, 500},
	//}
)

func StressTestCmd() *cobra.Command {
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

			gasLimit := uint64(cfg.Custom.GasLimit)
			fees := sdk.NewCoins(sdk.NewCoin(cfg.Custom.FeeDenom, sdk.NewInt(cfg.Custom.FeeAmount)))
			memo := cfg.Custom.Memo
			tx := tx.NewTransaction(client, chainID, gasLimit, fees, memo)

			f, err := os.OpenFile("result.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}
			defer f.Close()
			w := csv.NewWriter(f)
			fi, err := f.Stat()
			if err != nil {
				return fmt.Errorf("stat file: %w", err)
			}
			if fi.Size() == 0 {
				if err := w.Write([]string{
					"height",
					"num_broadcast_txs",
					"num_committed_txs",
					"planned_num_broadcast_txs",
				}); err != nil {
					return fmt.Errorf("emit header: %w", err)
				}
				w.Flush()
				if err := w.Error(); err != nil {
					return fmt.Errorf("write header: %w", err)
				}
			}

			d := NewAccountDispenser(client, cfg.Custom.Mnemonics)
			if err := d.Next(); err != nil {
				return fmt.Errorf("get next account: %w", err)
			}

			for no, scenario := range scenarios {
				st, err := client.RPC.Status(ctx)
				if err != nil {
					return fmt.Errorf("get status: %w", err)
				}
				startingHeight := st.SyncInfo.LatestBlockHeight + 2
				log.Info().Msgf("current block height is %d, waiting for the next block to be committed", st.SyncInfo.LatestBlockHeight)

				if err := rpcclient.WaitForHeight(client.RPC, startingHeight-1, nil); err != nil {
					return fmt.Errorf("wait for height: %w", err)
				}
				log.Info().Msgf("starting simulation #%d, rounds = %d, num txs per block = %d", no+1, scenario.Rounds, scenario.NumTxsPerBlock)

				targetHeight := startingHeight

				for i := 0; i < scenario.Rounds; i++ {
					st, err := client.RPC.Status(ctx)
					if err != nil {
						return fmt.Errorf("get status: %w", err)
					}
					if st.SyncInfo.LatestBlockHeight != targetHeight-1 {
						log.Warn().Int64("expected", targetHeight-1).Int64("got", st.SyncInfo.LatestBlockHeight).Msg("mismatching block height")
						targetHeight = st.SyncInfo.LatestBlockHeight + 1
					}

					started := time.Now()
					sent := 0
				loop:
					for sent < scenario.NumTxsPerBlock {
						msgs, err := tx.CreateSwapBot(ctx, d.Addr(), poolID, offerCoin, demandCoinDenom, 1)
						if err != nil {
							return fmt.Errorf("generate msgs: %s", err)
						}

						for sent < scenario.NumTxsPerBlock {
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
								if resp.TxResponse.Code == 0x14 {
									log.Warn().Msg("mempool is full, stopping")
									d.DecAccSeq()
									break loop
								}
								if resp.TxResponse.Code == 0x13 || resp.TxResponse.Code == 0x20 {
									if err := d.Next(); err != nil {
										return fmt.Errorf("get next account: %w", err)
									}
									log.Warn().Str("addr", d.Addr()).Uint64("seq", d.AccSeq()).Msgf("received %#v, using next account", resp.TxResponse)
									time.Sleep(500 * time.Millisecond)
									break
								} else {
									panic(fmt.Sprintf("%#v\n", resp.TxResponse))
								}
							}
							sent++
						}
					}
					log.Debug().Msgf("took %s broadcasting txs", time.Since(started))

					if err := rpcclient.WaitForHeight(client.RPC, targetHeight, nil); err != nil {
						return fmt.Errorf("wait for height: %w", err)
					}

					r, err := client.RPC.Block(ctx, &targetHeight)
					if err != nil {
						return err
					}
					log.Info().Int64("height", targetHeight).Int("broadcast-txs", sent).Int("committed-txs", len(r.Block.Txs)).Msg("block committed")

					if err := w.Write([]string{
						strconv.FormatInt(targetHeight, 10),
						strconv.Itoa(sent),
						strconv.Itoa(len(r.Block.Txs)),
						strconv.Itoa(scenario.NumTxsPerBlock),
					}); err != nil {
						return fmt.Errorf("emit row: %w", err)
					}
					w.Flush()
					if err := w.Error(); err != nil {
						return fmt.Errorf("write row: %w", err)
					}

					targetHeight++
				}

				started := time.Now()
				log.Debug().Msg("cooling down")
				for {
					st, err := client.RPC.NumUnconfirmedTxs(ctx)
					if err != nil {
						return fmt.Errorf("get status: %w", err)
					}
					if st.Total == 0 {
						break
					}
					time.Sleep(5 * time.Second)
				}
				log.Debug().Str("elapsed", time.Since(started).String()).Msg("done cooling down")
				time.Sleep(5 * time.Minute)
			}

			return nil
		},
	}
	return cmd
}
