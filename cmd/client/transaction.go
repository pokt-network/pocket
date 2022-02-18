package main

import (
	"encoding/hex"
	"math/big"
	"pocket/p2p/pre_p2p"
	"pocket/shared/crypto"
	"pocket/utility/types"
)

func NewSendTxBytes(state *pre_p2p.TestState) []byte {
	val1 := state.ValidatorMap[1]
	val2 := state.ValidatorMap[2]
	addr1, err := hex.DecodeString(val1.Address)
	if err != nil {
		panic(err)
	}
	addr2, err := hex.DecodeString(val2.Address)
	if err != nil {
		panic(err)
	}
	privateKey1, err := crypto.NewPrivateKey(val1.PrivateKey)
	if err != nil {
		panic(err)
	}
	amountBig := big.NewInt(1000000)
	amount := types.BigIntToString(amountBig)
	feeBig := big.NewInt(10000)
	fee := types.BigIntToString(feeBig)
	nonceBig := big.NewInt(0)
	nonce := types.BigIntToString(nonceBig)
	cdc := types.UtilityCodec()
	msg, err := cdc.ToAny(&types.MessageSend{
		FromAddress: addr1,
		ToAddress:   addr2,
		Amount:      amount,
	})
	if err != nil {
		panic(err)
	}
	tx := types.Transaction{
		Msg:       msg,
		Fee:       fee,
		Signature: nil,
		Nonce:     nonce,
	}
	err = tx.Sign(privateKey1)
	if err != nil {
		panic(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		panic(err)
	}
	return transactionBytes
}
