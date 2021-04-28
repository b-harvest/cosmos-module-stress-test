package tx_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/test-go/testify/require"

	"github.com/b-harvest/liquidity-stress-test/client"
	"github.com/b-harvest/liquidity-stress-test/config"
	"github.com/b-harvest/liquidity-stress-test/tx"
	"github.com/b-harvest/liquidity-stress-test/wallet"

	sdkclienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var (
	c   *client.Client
	cfg *config.Config

	rpcAddress  = "http://localhost:26657"
	grpcAddress = "localhost:9090"
)

func TestMain(m *testing.M) {
	c, _ = client.NewClient(rpcAddress, grpcAddress)

	cfg, _ = config.Read(config.DefaultConfigPath)

	os.Exit(m.Run())
}

func TestFindAllPairs(t *testing.T) {
	pairs := []struct {
		pairs []string
	}{
		{
			[]string{
				"uakt",
				"uatom",
				"ubtsg",
				"udvpn",
				"ugcyb",
				"uiris",
				"uluna",
				"ungm",
				"uxprt",
				"uxrn",
				"xrun",
			},
		},
	}

	for _, p := range pairs {
		for i := 0; i < len(p.pairs)-1; i++ {
			for j := i + 1; j < len(p.pairs); j++ {
				t.Log(p.pairs[i], p.pairs[j])
			}
		}
	}
}

func TestSendTxsByIncrementingSequence(t *testing.T) {
	mnemonic := "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"

	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(mnemonic, "")
	require.NoError(t, err)

	chainID, err := c.RPC.GetNetworkChainID(context.Background())
	require.NoError(t, err)

	depositCoins := sdktypes.NewCoins(sdktypes.NewCoin("uakt", sdktypes.NewInt(10_000_000)), sdktypes.NewCoin("uatom", sdktypes.NewInt(10_000_000)))

	msg, err := tx.MsgDeposit(accAddr, uint64(1), depositCoins)
	require.NoError(t, err)

	msgs := []sdktypes.Msg{msg}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 1; i++ {
		account, err := c.GRPC.GetBaseAccountInfo(ctx, accAddr)
		require.NoError(t, err)

		accSeq := account.GetSequence()
		accNum := account.GetAccountNumber()

		gasLimit := uint64(100000000)
		fees := sdktypes.NewCoins(sdktypes.NewCoin("stake", sdktypes.NewInt(0)))

		txBuilder := c.CliCtx.TxConfig.NewTxBuilder()
		txBuilder.SetMsgs(msgs...)
		txBuilder.SetGasLimit(gasLimit)
		txBuilder.SetFeeAmount(fees)
		txBuilder.SetMemo("")

		signMode := c.CliCtx.TxConfig.SignModeHandler().DefaultMode()

		for j := 0; j < 5; j++ {
			sigV2 := signing.SignatureV2{
				PubKey: privKey.PubKey(),
				Data: &signing.SingleSignatureData{
					SignMode:  signMode,
					Signature: nil,
				},
				Sequence: accSeq,
			}

			err = txBuilder.SetSignatures(sigV2)
			require.NoError(t, err)

			signerData := authsigning.SignerData{
				ChainID:       chainID,
				AccountNumber: accNum,
				Sequence:      accSeq,
			}

			sigV2, err = sdkclienttx.SignWithPrivKey(signMode, signerData, txBuilder, privKey, c.CliCtx.TxConfig, accSeq)
			require.NoError(t, err)

			err = txBuilder.SetSignatures(sigV2)
			require.NoError(t, err)

			accSeq = accSeq + 1

			txBytes, err := c.CliCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			resp, err := c.GRPC.BroadcastTx(ctx, txBytes)
			require.NoError(t, err)

			fmt.Println("AccSeq: ", accSeq)
			fmt.Println("AccNum: ", accNum)
			fmt.Println("Code: ", resp.TxResponse.Code)
			fmt.Println("Height: ", resp.TxResponse.Height)
			fmt.Println("TxHash: ", resp.TxResponse.TxHash)
			fmt.Println("")

			// time.Sleep(1 * time.Second)
		}
	}
}

func TestDepositWithinBatch(t *testing.T) {
	mnemonic := "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"

	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(mnemonic, "")
	require.NoError(t, err)

	chainID, err := c.RPC.GetNetworkChainID(context.Background())
	require.NoError(t, err)

	testCases := []struct {
		name         string
		accAddr      string
		privKey      *secp256k1.PrivKey
		poolId       uint64
		depositCoins sdktypes.Coins
	}{
		{
			"uakt/uatom",
			accAddr,
			privKey,
			uint64(1),
			sdktypes.NewCoins(sdktypes.NewCoin("uakt", sdktypes.NewInt(5000000)), sdktypes.NewCoin("uatom", sdktypes.NewInt(5000000))),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := tx.MsgDeposit(tc.accAddr, tc.poolId, tc.depositCoins)
			require.NoError(t, err)

			msgs := []sdktypes.Msg{msg}

			gasLimit := uint64(100000000)
			fees := sdktypes.NewCoins(sdktypes.NewCoin("stake", sdktypes.NewInt(0)))
			memo := ""

			tx := tx.NewTransaction(c, chainID, gasLimit, fees, memo)

			account, err := c.GRPC.GetBaseAccountInfo(context.Background(), accAddr)
			require.NoError(t, err)

			accSeq := account.GetSequence()
			accNum := account.GetAccountNumber()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			resp, err := tx.SignAndBroadcast(ctx, accSeq, accNum, privKey, msgs...)
			require.NoError(t, err)

			fmt.Println("Code: ", resp.TxResponse.Code)
			fmt.Println("Height: ", resp.TxResponse.Height)
			fmt.Println("TxHash: ", resp.TxResponse.TxHash)
		})
	}
}

func TestWithdrawWithinBatch(t *testing.T) {
	mnemonic := "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"

	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(mnemonic, "")
	require.NoError(t, err)

	chainID, err := c.RPC.GetNetworkChainID(context.Background())
	require.NoError(t, err)

	testCases := []struct {
		name     string
		accAddr  string
		privKey  *secp256k1.PrivKey
		poolId   uint64
		poolCoin sdktypes.Coin
	}{
		{
			"uakt/uatom",
			accAddr,
			privKey,
			uint64(1),
			sdktypes.NewCoin("pool94720F40B38D6DD93DCE184D264D4BE089EDF124A9C0658CDBED6CA18CF27752", sdktypes.NewInt(50)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := tx.MsgWithdraw(tc.accAddr, tc.poolId, tc.poolCoin)
			require.NoError(t, err)

			msgs := []sdktypes.Msg{msg}

			gasLimit := uint64(100000000)
			fees := sdktypes.NewCoins(sdktypes.NewCoin("stake", sdktypes.NewInt(0)))
			memo := ""

			tx := tx.NewTransaction(c, chainID, gasLimit, fees, memo)

			account, err := c.GRPC.GetBaseAccountInfo(context.Background(), accAddr)
			require.NoError(t, err)

			accSeq := account.GetSequence()
			accNum := account.GetAccountNumber()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			resp, err := tx.SignAndBroadcast(ctx, accSeq, accNum, privKey, msgs...)
			require.NoError(t, err)

			fmt.Println("Code: ", resp.TxResponse.Code)
			fmt.Println("Height: ", resp.TxResponse.Height)
			fmt.Println("TxHash: ", resp.TxResponse.TxHash)
		})
	}
}
