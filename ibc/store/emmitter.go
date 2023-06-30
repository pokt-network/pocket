package store

import (
	ibcTypes "github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
)

// emitUpdateStoreEvent handles an UpdateIBCStore event locally and then broadcasts it to the network
func emitUpdateStoreEvent(bus modules.Bus, key, value []byte) error {
	updateMsg := ibcTypes.CreateUpdateStoreMessage(key, value)
	anyUpdate, err := codec.GetCodec().ToAny(updateMsg)
	if err != nil {
		return err
	}
	if err := bus.GetIBCModule().HandleMessage(anyUpdate); err != nil {
		return err
	}

	// Broadcast event to the network
	letter, err := messaging.PackMessage(updateMsg)
	if err != nil {
		return err
	}
	anyLetter, err := codec.GetCodec().ToAny(letter)
	if err != nil {
		return err
	}
	if err := bus.GetP2PModule().Broadcast(anyLetter); err != nil {
		return err
	}

	return nil
}

// emitPruneStoreEvent handles an PruneIBCStore event locally and then broadcasts it to the network
func emitPruneStoreEvent(bus modules.Bus, key []byte) error {
	pruneMsg := ibcTypes.CreatePruneStoreMessage(key)
	anyPrune, err := codec.GetCodec().ToAny(pruneMsg)
	if err != nil {
		return err
	}
	if err := bus.GetIBCModule().HandleMessage(anyPrune); err != nil {
		return err
	}

	// Broadcast event to the network
	letter, err := messaging.PackMessage(pruneMsg)
	if err != nil {
		return err
	}
	anyLetter, err := codec.GetCodec().ToAny(letter)
	if err != nil {
		return err
	}
	if err := bus.GetP2PModule().Broadcast(anyLetter); err != nil {
		return err
	}

	return nil
}
