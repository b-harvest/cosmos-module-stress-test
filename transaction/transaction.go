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
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type Transaction struct {
	Client    *client.Client `json:"client"`
	ChainID   string         `json:"chain_id"`
	Mnemonic  string         `json:"mnemonic"`
	BondDenom string         `json:"bond_denom"`
}

// MsgCreatePool
func (t *Transaction) MsgCreatePool(ctx context.Context, accAddr string, depositCoinA, depositCoinB sdktypes.Coin) ([]byte, error) {
	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(t.Mnemonic, "")
	if err != nil {
		return nil, err
	}

	account, err := t.Client.GRPC.GetBaseAccountInfo(ctx, accAddr)
	accNumber := account.GetAccountNumber()
	accSeq := account.GetSequence()

	// msgs
	poolCreator, err := sdktypes.AccAddressFromBech32(accAddr)
	if err != nil {
		return nil, err
	}

	poolTypeId := uint32(1)
	depositCoins := sdktypes.NewCoins(depositCoinA, depositCoinB)

	msgCreatePool := liqtypes.NewMsgCreatePool(poolCreator, poolTypeId, depositCoins)
	msgCreatePool.ValidateBasic()

	msgs := []sdktypes.Msg{msgCreatePool}

	// fees
	fees := sdktypes.NewCoins(sdktypes.NewCoin(t.BondDenom, sdktypes.NewInt(0)))

	// memo
	memo := ""

	txBytes, err := t.Sign(msgs, fees, memo, accNumber, accSeq, privKey)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// MsgDeposit
func (t *Transaction) MsgDeposit(ctx context.Context, accAddr string, depositCoinA, depositCoinB sdktypes.Coin) ([]byte, error) {
	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(t.Mnemonic, "")
	if err != nil {
		return nil, err
	}

	account, err := t.Client.GRPC.GetBaseAccountInfo(ctx, accAddr)
	accNumber := account.GetAccountNumber()
	accSeq := account.GetSequence()

	// msgs
	poolCreator, err := sdktypes.AccAddressFromBech32(accAddr)
	if err != nil {
		return nil, err
	}

	poolId := uint64(1)
	depositCoins := sdktypes.NewCoins(depositCoinA, depositCoinB)

	msgDepositWithinBatch := liqtypes.NewMsgDepositWithinBatch(poolCreator, poolId, depositCoins)
	msgDepositWithinBatch.ValidateBasic()

	msgs := []sdktypes.Msg{msgDepositWithinBatch}

	// fees
	fees := sdktypes.NewCoins(sdktypes.NewCoin(t.BondDenom, sdktypes.NewInt(0)))

	// memo
	memo := ""

	txBytes, err := t.Sign(msgs, fees, memo, accNumber, accSeq, privKey)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// MsgWithdraw
func (t *Transaction) MsgWithdraw(ctx context.Context, poolCoin sdktypes.Coin) ([]byte, error) {
	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(t.Mnemonic, "")
	if err != nil {
		return nil, err
	}

	account, err := t.Client.GRPC.GetBaseAccountInfo(ctx, accAddr)
	accNumber := account.GetAccountNumber()
	accSeq := account.GetSequence()

	// msgs
	withdrawer, err := sdktypes.AccAddressFromBech32(accAddr)
	if err != nil {
		return nil, err
	}

	poolId := uint64(1)

	msgWithdrawWithinBatch := liqtypes.NewMsgWithdrawWithinBatch(withdrawer, poolId, poolCoin)
	msgWithdrawWithinBatch.ValidateBasic()

	msgs := []sdktypes.Msg{msgWithdrawWithinBatch}

	// fees
	fees := sdktypes.NewCoins(sdktypes.NewCoin(t.BondDenom, sdktypes.NewInt(0)))

	// memo
	memo := ""

	txBytes, err := t.Sign(msgs, fees, memo, accNumber, accSeq, privKey)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// MsgSwap
func (t *Transaction) MsgSwap() {

}

// WrapMultiMsgs
func WrapMultiMsgs() {

}

func (t *Transaction) Sign(msgs []sdktypes.Msg, fees sdktypes.Coins, memo string,
	accNumber uint64, accSeq uint64, privKey *secp256k1.PrivKey) ([]byte, error) {

	txBuilder := t.Client.CliCtx.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msgs...)
	txBuilder.SetGasLimit(200000)
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
