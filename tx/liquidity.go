package tx

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/nodebreaker0-0/cosmos-module-stress-test/client"

	liquiditytypes "github.com/tendermint/liquidity/x/liquidity/types"

	sdkclienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// Transaction is an object that has common fields when signing transaction.
type Transaction struct {
	Client   *client.Client `json:"client"`
	ChainID  string         `json:"chain_id"`
	GasLimit uint64         `json:"gas_limit"`
	Fees     sdktypes.Coins `json:"fees"`
	Memo     string         `json:"memo"`
}

// NewTransaction returns new Transaction object.
func NewTransaction(client *client.Client, chainID string, gasLimit uint64, fees sdktypes.Coins, memo string) *Transaction {
	return &Transaction{
		Client:   client,
		ChainID:  chainID,
		GasLimit: gasLimit,
		Fees:     fees,
		Memo:     memo,
	}
}

// MsgCreatePool creates create pool message and returns MsgCreatePool transaction message.
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

// MsgDeposit creates deposit message and returns MsgDeposit transaction message.
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

// MsgWithdraw creates withdraw message and returns MsgWithdraw transaction message.
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

// MsgSwap creates swap message and returns MsgWithdraw MsgSwap message.
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

// CreateSwapBot creates a bot that makes multiple swaps which increases and decreases
func (t *Transaction) CreateSwapBot(ctx context.Context, poolCreator string, poolId uint64, offerCoin sdktypes.Coin, msgNum int) ([]sdktypes.Msg, error) {
	pool, err := t.Client.GRPC.GetPool(ctx, poolId)
	if err != nil {
		return []sdktypes.Msg{}, err
	}

	reserveCoins := sdktypes.NewCoins()

	for _, denom := range pool.ReserveCoinDenoms {
		coin, err := t.Client.GRPC.GetBalance(ctx, pool.GetReserveAccount().String(), denom)
		if err != nil {
			return []sdktypes.Msg{}, err
		}
		reserveCoins = reserveCoins.Add(*coin)
	}

	orderPrice := reserveCoins.AmountOf(pool.ReserveCoinDenoms[0]).ToDec().Quo(reserveCoins.AmountOf(pool.ReserveCoinDenoms[1]).ToDec())

	var msgs []sdktypes.Msg

	// randomize order price
	for i := 0; i < msgNum; i++ {
		random := sdktypes.NewDec(int64(rand.Intn(2)))
		orderPricePercentage := orderPrice.Mul(random.Quo(sdktypes.NewDec(100)))

		if i%2 == 0 {
			orderPrice = orderPrice.Add(orderPricePercentage)
		} else {
			orderPrice = orderPrice.Sub(orderPricePercentage)
		}

		msg, err := MsgSwap(poolCreator, poolId, uint32(1), offerCoin, pool.ReserveCoinDenoms[1], orderPrice, sdktypes.NewDecWithPrec(3, 3))
		if err != nil {
			return []sdktypes.Msg{}, err
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// Sign signs message(s) with the account's private key and braodacasts the message(s).
func (t *Transaction) Sign(ctx context.Context, accSeq uint64, accNum uint64, privKey *secp256k1.PrivKey, msgs ...sdktypes.Msg) ([]byte, error) {
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
