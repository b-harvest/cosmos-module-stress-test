package tx

import (
	"context"
	"fmt"

	"github.com/nodebreaker0-0/cosmos-module-stress-test/client"
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdkclienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channelutils "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/client/utils"
)

// Transaction is an object that has common fields when signing transaction.
type Ibctransaction struct {
	Client   *client.Client `json:"client"`
	ChainID  string         `json:"chain_id"`
	GasLimit uint64         `json:"gas_limit"`
	Fees     sdktypes.Coins `json:"fees"`
	Memo     string         `json:"memo"`
}

const (
	flagPacketTimeoutHeight    = "packet-timeout-height"
	flagPacketTimeoutTimestamp = "packet-timeout-timestamp"
	flagAbsoluteTimeouts       = "absolute-timeouts"
)

// NewTransaction returns new Transaction object.
func IbcNewtransaction(client *client.Client, chainID string, gasLimit uint64, fees sdktypes.Coins, memo string) *Transaction {
	return &Transaction{
		Client:   client,
		ChainID:  chainID,
		GasLimit: gasLimit,
		Fees:     fees,
		Memo:     memo,
	}
}

// MsgCreatePool creates create pool message and returns MsgCreatePool transaction message.
func MsgTransfer(cmd *cobra.Command, ctx sdkclient.Context, srcPort string, srcChannel string, coin sdktypes.Coin, sender string, receiver string) (sdktypes.Msg, error) {
	ibcsender, err := sdktypes.AccAddressFromBech32(sender)
	if err != nil {
		return &ibctypes.MsgTransfer{}, err
	}
	timeoutHeightStr, err := cmd.Flags().GetString(flagPacketTimeoutHeight)
	if err != nil {
		return nil, err
	}
	timeoutHeight, err := clienttypes.ParseHeight(timeoutHeightStr)
	if err != nil {
		return nil, err
	}

	timeoutTimestamp, err := cmd.Flags().GetUint64(flagPacketTimeoutTimestamp)
	if err != nil {
		return nil, err
	}

	absoluteTimeouts, err := cmd.Flags().GetBool(flagAbsoluteTimeouts)
	if err != nil {
		return nil, err
	}

	if !absoluteTimeouts {
		consensusState, height, _, err := channelutils.QueryLatestConsensusState(ctx, srcPort, srcChannel)
		if err != nil {
			return nil, err
		}

		if !timeoutHeight.IsZero() {
			absoluteHeight := height
			absoluteHeight.RevisionNumber += timeoutHeight.RevisionNumber
			absoluteHeight.RevisionHeight += timeoutHeight.RevisionHeight
			timeoutHeight = absoluteHeight
		}

		if timeoutTimestamp != 0 {
			timeoutTimestamp = consensusState.GetTimestamp() + timeoutTimestamp
		}
	}
	msg := ibctypes.NewMsgTransfer(srcPort, srcChannel, coin, ibcsender, receiver, timeoutHeight, timeoutTimestamp)

	if err := msg.ValidateBasic(); err != nil {
		return &ibctypes.MsgTransfer{}, err
	}

	return msg, nil
}

func (t *Transaction) CreateTransferBot(cmd *cobra.Command, ctx sdkclient.Context, srcPort string, srcChannel string, coin sdktypes.Coin, sender string, receiver string, msgNum int) ([]sdktypes.Msg, error) {

	var msgs []sdktypes.Msg

	for i := 0; i < msgNum; i++ {

		msg, err := MsgTransfer(cmd, ctx, srcPort, srcChannel, coin, sender, receiver)
		if err != nil {
			return []sdktypes.Msg{}, err
		}
		msgs = append(msgs, msg)

	}
	return msgs, nil
}

// Sign signs message(s) with the account's private key and braodacasts the message(s).
func (t *Transaction) IbcSign(ctx context.Context, accSeq uint64, accNum uint64, privKey *secp256k1.PrivKey, msgs ...sdktypes.Msg) ([]byte, error) {
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
		return nil, fmt.Errorf("failed to set signatures: %s", err)
	}

	signerData := authsigning.SignerData{
		ChainID:       t.ChainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}

	sigV2, err = sdkclienttx.SignWithPrivKey(signMode, signerData, txBuilder, privKey, t.Client.CliCtx.TxConfig, accSeq)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with private key: %s", err)
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signatures: %s", err)
	}

	txByte, err := t.Client.CliCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode tx and get raw tx data: %s", err)
	}

	return txByte, nil
}
