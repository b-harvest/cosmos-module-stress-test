package transaction

import (
	"context"
	"fmt"

	"github.com/b-harvest/liquidity-stress-test/client"
	"github.com/b-harvest/liquidity-stress-test/wallet"

	liqtypes "github.com/tendermint/liquidity/x/liquidity/types"

	sdkclienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/rs/zerolog/log"
)

// Transaction is an object that has common fields for signing a transaction.
type Transaction struct {
	Client   *client.Client `json:"client"`
	ChainID  string         `json:"chain_id"`
	Mnemonic string         `json:"mnemonic"`
	FeeDenom string         `json:"fee_denom"`
}

// NewTransaction returns new Transaction.
func NewTransaction(client *client.Client, chainID string, mnemonic string, feeDenom string) *Transaction {
	return &Transaction{
		Client:   client,
		ChainID:  chainID,
		Mnemonic: mnemonic,
		FeeDenom: feeDenom,
	}
}

// SignAndBroadcastMsgCreatePool wraps MsgCreatePool and signs it with account's signature using its private key.
func (t *Transaction) SignAndBroadcastMsgCreatePool(ctx context.Context, depositCoinA, depositCoinB sdktypes.Coin) (*tx.BroadcastTxResponse, error) {
	log.Debug().Msg("signing MsgCreatePool and broadcast raw tx data")

	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(t.Mnemonic, "")
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	account, err := t.Client.GRPC.GetBaseAccountInfo(ctx, accAddr)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	accNumber := account.GetAccountNumber()
	accSeq := account.GetSequence()

	// msgs
	poolCreator, err := sdktypes.AccAddressFromBech32(accAddr)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	poolTypeId := uint32(1)
	depositCoins := sdktypes.NewCoins(depositCoinA, depositCoinB)

	msgCreatePool := liqtypes.NewMsgCreatePool(poolCreator, poolTypeId, depositCoins)
	msgCreatePool.ValidateBasic()

	msgs := []sdktypes.Msg{msgCreatePool}

	// fees
	fees := sdktypes.NewCoins(sdktypes.NewCoin(t.FeeDenom, sdktypes.NewInt(0)))

	// memo
	memo := ""

	txBytes, err := t.sign(msgs, fees, memo, accNumber, accSeq, privKey)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	res, err := t.Client.GRPC.BroadcastTx(txBytes)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	return res, nil
}

// SignAndBroadcastMsgDeposit wraps MsgDeposit and signs it with account's signature using its private key.
func (t *Transaction) SignAndBroadcastMsgDeposit(ctx context.Context, depositCoinA, depositCoinB sdktypes.Coin) (*tx.BroadcastTxResponse, error) {
	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(t.Mnemonic, "")
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	account, err := t.Client.GRPC.GetBaseAccountInfo(ctx, accAddr)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	accNumber := account.GetAccountNumber()
	accSeq := account.GetSequence()

	// msgs
	poolCreator, err := sdktypes.AccAddressFromBech32(accAddr)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	poolId := uint64(1)
	depositCoins := sdktypes.NewCoins(depositCoinA, depositCoinB)

	msgDepositWithinBatch := liqtypes.NewMsgDepositWithinBatch(poolCreator, poolId, depositCoins)
	msgDepositWithinBatch.ValidateBasic()

	msgs := []sdktypes.Msg{msgDepositWithinBatch}

	// fees
	fees := sdktypes.NewCoins(sdktypes.NewCoin(t.FeeDenom, sdktypes.NewInt(0)))

	// memo
	memo := ""

	txBytes, err := t.sign(msgs, fees, memo, accNumber, accSeq, privKey)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	res, err := t.Client.GRPC.BroadcastTx(txBytes)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	return res, nil
}

// SignAndBroadcastMsgWithdraw wraps MsgWithdraw and signs it with account's signature using its private key.
func (t *Transaction) SignAndBroadcastMsgWithdraw(ctx context.Context, poolCoin sdktypes.Coin) (*tx.BroadcastTxResponse, error) {
	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(t.Mnemonic, "")
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	account, err := t.Client.GRPC.GetBaseAccountInfo(ctx, accAddr)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	accNumber := account.GetAccountNumber()
	accSeq := account.GetSequence()

	// msgs
	withdrawer, err := sdktypes.AccAddressFromBech32(accAddr)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	poolId := uint64(1)

	msgWithdrawWithinBatch := liqtypes.NewMsgWithdrawWithinBatch(withdrawer, poolId, poolCoin)
	msgWithdrawWithinBatch.ValidateBasic()

	msgs := []sdktypes.Msg{msgWithdrawWithinBatch}

	// fees
	fees := sdktypes.NewCoins(sdktypes.NewCoin(t.FeeDenom, sdktypes.NewInt(0)))

	// memo
	memo := ""

	txBytes, err := t.sign(msgs, fees, memo, accNumber, accSeq, privKey)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	res, err := t.Client.GRPC.BroadcastTx(txBytes)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	return res, nil
}

// SignAndBroadcastMsgSwap wraps MsgSwap and signs it with account's signature using its private key.
func (t *Transaction) SignAndBroadcastMsgSwap(ctx context.Context, offerCoin sdktypes.Coin, demandCoinDenom string,
	orderPrice sdktypes.Dec, swapFeeRate sdktypes.Dec) (*tx.BroadcastTxResponse, error) {

	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(t.Mnemonic, "")
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	account, err := t.Client.GRPC.GetBaseAccountInfo(ctx, accAddr)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	accNumber := account.GetAccountNumber()
	accSeq := account.GetSequence()

	// msgs
	swapRequester, err := sdktypes.AccAddressFromBech32(accAddr)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	poolId := uint64(1)
	swapTypeId := uint32(1)

	msgWithdrawWithinBatch := liqtypes.NewMsgSwapWithinBatch(swapRequester, poolId, swapTypeId, offerCoin, demandCoinDenom, orderPrice, swapFeeRate)
	msgWithdrawWithinBatch.ValidateBasic()

	msgs := []sdktypes.Msg{msgWithdrawWithinBatch}

	// fees
	fees := sdktypes.NewCoins(sdktypes.NewCoin(t.FeeDenom, sdktypes.NewInt(0)))

	// memo
	memo := ""

	txBytes, err := t.sign(msgs, fees, memo, accNumber, accSeq, privKey)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	res, err := t.Client.GRPC.BroadcastTx(txBytes)
	if err != nil {
		return &tx.BroadcastTxResponse{}, err
	}

	return res, nil
}

// WrapMultiMsgs
func WrapMultiMsgs() {

}

// sign generates signatures and provide canonical bytes to sign over.
func (t *Transaction) sign(msgs []sdktypes.Msg, fees sdktypes.Coins, memo string,
	accNumber uint64, accSeq uint64, privKey *secp256k1.PrivKey) ([]byte, error) {

	txBuilder := t.Client.CliCtx.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msgs...)
	txBuilder.SetGasLimit(10000000)
	txBuilder.SetFeeAmount(fees)
	txBuilder.SetMemo(memo)

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
		return nil, fmt.Errorf("failed to set signatures: %s", err)
	}

	// signatures
	signerData := authsigning.SignerData{
		ChainID:       t.ChainID,
		AccountNumber: accNumber,
		Sequence:      accSeq,
	}

	if sigV2, err = sdkclienttx.SignWithPrivKey(signMode, signerData, txBuilder, privKey, t.Client.CliCtx.TxConfig, accSeq); err != nil {
		return nil, fmt.Errorf("failed to sign with private key: %s", err)
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signatures: %s", err)
	}

	return t.Client.CliCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
}
