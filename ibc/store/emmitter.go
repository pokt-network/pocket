package store

import (
	"encoding/hex"

	ibcTypes "github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility"
)

// emitUpdateStoreEvent handles an UpdateIBCStore event locally and then broadcasts it to the network
func emitUpdateStoreEvent(bus modules.Bus, privateKey string, key, value []byte) error {
	updateMsg := ibcTypes.CreateUpdateStoreMessage(key, value)
	updateTx, err := ibcTypes.ConvertIBCMessageToTx(updateMsg)
	if err != nil {
		return err
	}
	return broadcastEvent(bus, updateTx, privateKey)
}

// emitPruneStoreEvent handles an PruneIBCStore event locally and then broadcasts it to the network
func emitPruneStoreEvent(bus modules.Bus, privateKey string, key []byte) error {
	pruneMsg := ibcTypes.CreatePruneStoreMessage(key)
	pruneTx, err := ibcTypes.ConvertIBCMessageToTx(pruneMsg)
	if err != nil {
		return err
	}
	return broadcastEvent(bus, pruneTx, privateKey)
}

// broadcastEvent handles an IBC event locally and then broadcasts it to the network
func broadcastEvent(bus modules.Bus, tx *coreTypes.Transaction, privateKey string) error {
	txBz, err := signAndMarshal(tx, privateKey)
	if err != nil {
		return err
	}
	if err := bus.GetUtilityModule().HandleTransaction(txBz); err != nil {
		return err
	}
	utilityMsg, err := utility.PrepareTxGossipMessage(txBz)
	if err != nil {
		return err
	}
	if err := bus.GetP2PModule().Broadcast(utilityMsg); err != nil {
		return err
	}
	return nil
}

// signAndMarshal signs the transaction with the private key provided and returns the
// serialised signed transaction
func signAndMarshal(tx *coreTypes.Transaction, privateKey string) ([]byte, error) {
	pkBz, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}
	pk, err := crypto.NewPrivateKeyFromBytes(pkBz)
	if err != nil {
		return nil, err
	}
	signableBz, err := tx.SignableBytes()
	if err != nil {
		return nil, err
	}
	signature, err := pk.Sign(signableBz)
	if err != nil {
		return nil, err
	}
	tx.Signature = &coreTypes.Signature{
		Signature: signature,
		PublicKey: pk.PublicKey().Bytes(),
	}
	txBz, err := codec.GetCodec().Marshal(tx)
	if err != nil {
		return nil, err
	}
	return txBz, nil
}
