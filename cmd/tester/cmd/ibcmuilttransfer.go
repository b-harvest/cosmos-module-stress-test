package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/b-harvest/cosmos-module-stress-test/client"
	"github.com/b-harvest/cosmos-module-stress-test/config"
	"github.com/b-harvest/cosmos-module-stress-test/query"
	"github.com/b-harvest/cosmos-module-stress-test/tx"
	"github.com/b-harvest/cosmos-module-stress-test/wallet"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"

	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

func IBCMuiltTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "muilt-transfer [src-chains] [dst-chains] [amount] [blocks] [tx-num] [msg-num]",
		Short:   "muilt Transfer a fungible token through IBC",
		Aliases: []string{"mt"},
		Args:    cobra.ExactArgs(6),
		Long: `Transfer a fungible token through IBC.

Example: $tester mt gaia,iris terra,osmo 10 1 1 1

#Only chains defined in ibccconfig are available.
#As many chain numbers as defined in src-chains, mnemonics are required.

src-chains: Group of chains sending tokens.
dst-chains: The group of chains that receive tokens. Receive tokens from all chains defined in src-chains.
blocks: how many blocks to keep the test going?
tx-num: how many transactions to be included in a block
msg-num: how many transaction messages to be included in a transaction
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := SetLogger(logLevel)
			if err != nil {
				return err
			}

			cfg, err := config.Read(config.DefaultConfigPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %s", err)
			}

			srcchains := strings.Split(args[0], ",")
			for _, srcchain := range srcchains {
				var flag bool
				for _, chain := range cfg.IBCconfig.Chains {
					flag = true
					if chain.ChainId == srcchain {
						break
					}
					flag = false
				}
				if !flag {
					return fmt.Errorf("the entered src chains does not exist in config")
				}
			}

			dstchains := strings.Split(args[1], ",")
			for _, dstchain := range dstchains {
				var flag bool
				for _, chain := range cfg.IBCconfig.Chains {
					flag = true
					if chain.ChainId == dstchain {
						break
					}
					flag = false
				}
				if !flag {
					return fmt.Errorf("the entered dst chains does not exist in config")
				}
			}
			DstchainsSize := len(dstchains)
			MnemonicsSize := len(cfg.Custom.Mnemonics)
			if DstchainsSize > MnemonicsSize {
				return fmt.Errorf("the number of ibcconfig and mnemics is different")
			}
			var wg sync.WaitGroup
			for _, chainname := range srcchains {
				wg.Add(1)
				go func(chainname string) error {
					defer wg.Done()
					var mainchain config.IBCchain
					var subchains []config.IBCchain

					for _, ibcconfigchain := range cfg.IBCconfig.Chains {
						if ibcconfigchain.ChainId == chainname {
							mainchain = ibcconfigchain
							break
						}
					}

					for _, chainname := range dstchains {
						for _, ibcconfigchain := range cfg.IBCconfig.Chains {
							if chainname == mainchain.ChainId {
								break
							}
							if ibcconfigchain.ChainId == chainname {
								subchains = append(subchains, ibcconfigchain)
								break
							}
						}
					}

					mainchainibcinfo, err := query.AllChainsTrace(mainchain.Grpc)
					if err != nil {
						return err
					}

					for index, dstchaininfo := range subchains {
						wg.Add(1)
						go func(index int, dstchaininfo config.IBCchain) error {
							defer wg.Done()
							err := DstChainsend(cmd, index, dstchaininfo, mainchainibcinfo, mainchain, cfg, args)
							if err != nil {
								return err
							}
							return err
						}(index, dstchaininfo)
					} //src chain-> dst chains send for loop

					return err
				}(chainname)
			} //src chains for loop
			wg.Wait()
			return nil
		},
	}
	cmd.Flags().String(flagPacketTimeoutHeight, ibctypes.DefaultRelativePacketTimeoutHeight, "Packet timeout block height. The timeout is disabled when set to 0-0.")
	cmd.Flags().Uint64(flagPacketTimeoutTimestamp, ibctypes.DefaultRelativePacketTimeoutTimestamp, "Packet timeout timestamp in nanoseconds. Default is 10 minutes. The timeout is disabled when set to 0.")
	cmd.Flags().Bool(flagAbsoluteTimeouts, false, "Timeout flags are used as absolute timeouts.")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func DstChainsend(cmd *cobra.Command, accountindex int, dstchaininfo config.IBCchain, mainchainibcinfo []query.ClientIds, mainchain config.IBCchain, cfg *config.Config, args []string) error {
	client, err := client.NewClient(mainchain.Rpc, mainchain.Grpc)
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

	var srcPort string
	var srcChannel string
	var receiver string
	for _, i := range mainchainibcinfo {
		if dstchaininfo.ChainId == i.ClientChainId {
			srcPort = "transfer"
			for _, j := range i.ConnectIDs {
				srcChannel = j.ChannsIDs[0]
				receiver = dstchaininfo.DstAddress
			}
			break
		}
	}
	sendcoin := args[2] + mainchain.TokenDenom
	coin, err := sdktypes.ParseCoinNormalized(sendcoin)
	if err != nil {
		return err
	}
	accAddr, privKey, err := wallet.IBCRecoverAccountFromMnemonic(cfg.Custom.Mnemonics[accountindex], "", mainchain.AccountHD, mainchain.AccountaddrPrefix)
	if err != nil {
		return fmt.Errorf("failed to retrieve account from mnemonic: %s", err)
	}
	if !strings.HasPrefix(coin.Denom, "ibc/") {
		denomTrace := ibctypes.ParseDenomTrace(coin.Denom)
		coin.Denom = denomTrace.IBCDenom()
	}

	blocks, err := strconv.Atoi(args[3])
	if err != nil {
		return fmt.Errorf("blocks must be integer: %s", args[0])
	}

	txNum, err := strconv.Atoi(args[4])
	if err != nil {
		return fmt.Errorf("txNum must be integer: %s", args[0])
	}

	msgNum, err := strconv.Atoi(args[5])
	if err != nil {
		return fmt.Errorf("txNum must be integer: %s", args[0])
	}

	gasLimit := uint64(cfg.Custom.GasLimit)
	fees := sdktypes.NewCoins(sdktypes.NewCoin(mainchain.TokenDenom, sdktypes.NewInt(cfg.Custom.FeeAmount)))
	memo := cfg.Custom.Memo
	tx := tx.IbcNewtransaction(client, chainID, gasLimit, fees, memo)
	account, err := client.GRPC.GetBaseAccountInfo(ctx, accAddr)
	if err != nil {
		return fmt.Errorf("failed to get account information: %s", err)
	}
	accSeq := account.GetSequence()
	accNum := account.GetAccountNumber()
	blockTimes := make(map[int64]time.Time)
	st, err := client.RPC.Status(ctx)
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}
	startingHeight := st.SyncInfo.LatestBlockHeight + 2
	log.Info().Msgf("current block height is %d, waiting for the next block to be committed <%s>", st.SyncInfo.LatestBlockHeight, mainchain.ChainId)

	if err := rpcclient.WaitForHeight(client.RPC, startingHeight-1, nil); err != nil {
		return fmt.Errorf("wait for height: %w", err)
	}
	log.Info().Msgf("starting simulation #%d, blocks = %d, num txs per block = %d <%s>", blocks+1, blocks, txNum, mainchain.ChainId)
	targetHeight := startingHeight

	for i := 0; i < blocks; i++ {
		st, err := client.RPC.Status(ctx)
		if err != nil {
			return fmt.Errorf("get status: %w", err)
		}
		if st.SyncInfo.LatestBlockHeight != targetHeight-1 {
			log.Warn().Int64("expected", targetHeight-1).Int64("got", st.SyncInfo.LatestBlockHeight).Msg("mismatching block height")
			targetHeight = st.SyncInfo.LatestBlockHeight + 1
		}

		//started := time.Now()
		sent := 0
	loop:
		for sent < txNum {
			msgs, err := tx.CreateTransferBot(cmd, ibcclientCtx, srcPort, srcChannel, coin, accAddr, receiver, msgNum)
			if err != nil {
				return fmt.Errorf("failed to create msg: %s", err)
			}
			for sent < txNum {
				txByte, err := tx.IbcSign(ctx, accSeq, accNum, privKey, msgs...)
				if err != nil {
					return fmt.Errorf("failed to sign and broadcast: %s", err)
				}
				resp, err := client.GRPC.BroadcastTx(ctx, txByte)
				//log.Info().Msgf("took %s broadcasting txs", resp)
				if err != nil {
					return fmt.Errorf("broadcast tx: %w", err)
				}
				accSeq = accSeq + 1
				if resp.TxResponse.Code != 0 {
					if resp.TxResponse.Code == 0x14 {
						log.Warn().Msg("mempool is full, stopping")
						accSeq = accSeq - 1
						break loop
					}
				}
				sent++
			}
		}
		//log.Debug().Msgf("took %s broadcasting txs", time.Since(started))

		if err := rpcclient.WaitForHeight(client.RPC, targetHeight, nil); err != nil {
			return fmt.Errorf("wait for height: %w", err)
		}
		r, err := client.RPC.Block(ctx, &targetHeight)
		if err != nil {
			return err
		}
		var blockDuration time.Duration
		bt, ok := blockTimes[targetHeight-1]
		if !ok {
			log.Warn().Msg("past block time not found")
		} else {
			blockDuration = r.Block.Time.Sub(bt)
			delete(blockTimes, targetHeight-1)
		}
		blockTimes[targetHeight] = r.Block.Time
		log.Info().
			Int64("height", targetHeight).
			Str("srcchain", mainchain.ChainId).
			Str("dstchain", dstchaininfo.ChainId).
			Str("block-time", r.Block.Time.Format(time.RFC3339Nano)).
			Str("block-duration", blockDuration.String()).
			Int("broadcast-txs", sent).
			Int("committed-txs", len(r.Block.Txs)).
			Msg("block committed")
		targetHeight++
	}
	return nil
}
