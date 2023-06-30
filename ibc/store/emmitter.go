package store

import (
	ibcTypes "github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
)

// emitUpdateStoreEvent emits an UpdateIBCStore event to the local bus and broadcasts it to the network
func emitUpdateStoreEvent(bus modules.Bus, key, value []byte) error {
	updateMsg := ibcTypes.CreateUpdateStoreMessage(key, value)
	letter, err := messaging.PackMessage(updateMsg)
	if err != nil {
		return err
	}
	// Publish the message to the bus to be handled locally
	bus.PublishEventToBus(letter)
	// Broadcast event to the network
	anyLetter, err := codec.GetCodec().ToAny(letter)
	if err != nil {
		return err
	}
	if err := bus.GetP2PModule().Broadcast(anyLetter); err != nil {
		return err
	}
	return nil
}

// emitDeleteStoreEvent emits a PruneIBCStore event to the local bus and broadcasts it to the network
func emitPruneStoreEvent(bus modules.Bus, key []byte) error {
	pruneMsg := ibcTypes.CreatePruneStoreMessage(key)
	letter, err := messaging.PackMessage(pruneMsg)
	if err != nil {
		return err
	}
	// Publish the message to the bus to be handled locally
	bus.PublishEventToBus(letter)
	// Broadcast event to the network
	anyLetter, err := codec.GetCodec().ToAny(letter)
	if err != nil {
		return err
	}
	if err := bus.GetP2PModule().Broadcast(anyLetter); err != nil {
		return err
	}
	return nil
}
