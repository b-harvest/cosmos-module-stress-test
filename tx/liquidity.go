package tx

import (
	"fmt"

	"github.com/b-harvest/liquidity-stress-test/client"

	liquiditytypes "github.com/tendermint/liquidity/x/liquidity/types"

	sdkclienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/rs/zerolog/log"
)

var (
	DefaultCreatePoolDepositCoinA = sdktypes.NewInt(50_000_000_000)
	DefaultCreatePoolDepositCoinB = sdktypes.NewInt(50_000_000_000)
	DefaultDepositCoinA           = sdktypes.NewInt(5000000)
	DefaultDepositCoinB           = sdktypes.NewInt(5000000)
	DefaultWithdrawPoolCoinA      = sdktypes.NewInt(50)
	DefaultSwapOfferCoin          = sdktypes.NewInt(50_000_000)

	DefaultFees     = sdktypes.NewCoins(sdktypes.NewCoin("stake", sdktypes.NewInt(0)))
	DefaultGasLimit = uint64(100000000)
	DefaultMemo     = ""
)

// Transaction is an object that has common fields for signing a transaction.
type Transaction struct {
	Client   *client.Client `json:"client"`
	ChainID  string         `json:"chain_id"`
	GasLimit uint64         `json:"gas_limit"`
	Fees     sdktypes.Coins `json:"fees"`
	Memo     string         `json:"memo"`
}

// NewTransaction returns new Transaction.
func NewTransaction(client *client.Client, chainID string, gasLimit uint64, fees sdktypes.Coins, memo string) *Transaction {
	return &Transaction{
		Client:   client,
		ChainID:  chainID,
		GasLimit: gasLimit,
		Fees:     fees,
		Memo:     memo,
	}
}

// MsgCreatePool returns MsgCreatePool object.
func MsgCreatePool(poolCreator string, poolTypeId uint32, depositCoins sdktypes.Coins) (sdktypes.Msg, error) {
	accAddr, err := sdktypes.AccAddressFromBech32(poolCreator)
	if err != nil {
		return &liquiditytypes.MsgCreatePool{}, err
	}

	msg := liquiditytypes.NewMsgCreatePool(accAddr, poolTypeId, depositCoins)

	if err := msg.ValidateBasic(); err != nil {
		return &liquiditytypes.MsgCreatePool{}, err
	}

	return msg, nil
}

// MsgDeposit returns MsgDeposit object.
func MsgDeposit(poolCreator string, poolId uint64, depositCoins sdktypes.Coins) (sdktypes.Msg, error) {
	accAddr, err := sdktypes.AccAddressFromBech32(poolCreator)
	if err != nil {
		return &liquiditytypes.MsgDepositWithinBatch{}, err
	}

	msg := liquiditytypes.NewMsgDepositWithinBatch(accAddr, poolId, depositCoins)

	if err := msg.ValidateBasic(); err != nil {
		return &liquiditytypes.MsgDepositWithinBatch{}, err
	}

	return msg, nil
}

// MsgWithdraw returns MsgWithdraw object.
func MsgWithdraw(poolCreator string, poolId uint64, poolCoin sdktypes.Coin) (sdktypes.Msg, error) {
	accAddr, err := sdktypes.AccAddressFromBech32(poolCreator)
	if err != nil {
		return &liquiditytypes.MsgWithdrawWithinBatch{}, err
	}

	msg := liquiditytypes.NewMsgWithdrawWithinBatch(accAddr, poolId, poolCoin)

	if err := msg.ValidateBasic(); err != nil {
		return &liquiditytypes.MsgWithdrawWithinBatch{}, err
	}

	return msg, nil
}

// MsgSwap returns MsgSwap object.
func MsgSwap(poolCreator string, poolId uint64, swapTypeId uint32, offerCoin sdktypes.Coin,
	demandCoinDenom string, orderPrice sdktypes.Dec, swapFeeRate sdktypes.Dec) (sdktypes.Msg, error) {
	accAddr, err := sdktypes.AccAddressFromBech32(poolCreator)
	if err != nil {
		return &liquiditytypes.MsgSwapWithinBatch{}, err
	}

	msg := liquiditytypes.NewMsgSwapWithinBatch(accAddr, poolId, swapTypeId, offerCoin, demandCoinDenom, orderPrice, swapFeeRate)

	if err := msg.ValidateBasic(); err != nil {
		return &liquiditytypes.MsgSwapWithinBatch{}, err
	}

	return msg, nil
}

// SignAndBroadcast signs message(s) with the account's private key and braodacasts the message(s).
func (t *Transaction) SignAndBroadcast(accSeq uint64, accNum uint64, privKey *secp256k1.PrivKey, msgs ...sdktypes.Msg) (*tx.BroadcastTxResponse, error) {
	txBuilder := t.Client.CliCtx.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msgs...)
	txBuilder.SetGasLimit(t.GasLimit)
	txBuilder.SetFeeAmount(t.Fees)
	txBuilder.SetMemo(t.Memo)

	signMode := t.Client.CliCtx.TxConfig.SignModeHandler().DefaultMode()

	sigV2 := signing.SignatureV2{
		PubKey: privKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: nil,
		},
		Sequence: accSeq,
	}

	err := txBuilder.SetSignatures(sigV2)
	if err != nil {
		return &tx.BroadcastTxResponse{}, fmt.Errorf("failed to set signatures: %s", err)
	}

	signerData := authsigning.SignerData{
		ChainID:       t.ChainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}

	sigV2, err = sdkclienttx.SignWithPrivKey(signMode, signerData, txBuilder, privKey, t.Client.CliCtx.TxConfig, accSeq)
	if err != nil {
		return &tx.BroadcastTxResponse{}, fmt.Errorf("failed to sign with private key: %s", err)
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return &tx.BroadcastTxResponse{}, fmt.Errorf("failed to set signatures: %s", err)
	}

	txBytes, err := t.Client.CliCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return &tx.BroadcastTxResponse{}, fmt.Errorf("failed to encode tx and get raw tx data: %s", err)
	}

	log.Debug().Msg("broadcasting transaction")

	resp, err := t.Client.GRPC.BroadcastTx(txBytes)
	if err != nil {
		return &tx.BroadcastTxResponse{}, fmt.Errorf("failed to broadcast transaction: %s", err)
	}

	return resp, nil
}
