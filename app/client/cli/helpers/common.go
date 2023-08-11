package helpers

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
)

// TECHDEBT: Accept reading this from `Datadir` and/or as a flag.
var genesisPath = runtime.GetEnv("GENESIS_PATH", "build/config/genesis.json")

// FetchPeerstore retrieves the providers from the CLI context and uses them to retrieve the address book for the current height
func FetchPeerstore(cmd *cobra.Command) (types.Peerstore, error) {
	bus, err := GetBusFromCmd(cmd)
	if err != nil {
		return nil, err
	}
	// TECHDEBT(#811): Remove type casting once Peerstore is available as a submodule
	pstoreProvider, err := peerstore_provider.GetPeerstoreProvider(bus)
	if err != nil {
		return nil, err
	}
	pstore, err := pstoreProvider.GetStakedPeerstoreAtCurrentHeight()
	if err != nil {
		return nil, fmt.Errorf("retrieving peerstore")
	}
	return pstore, nil
}

// sendConsensusNewHeightEventToP2PModule mimicks the consensus module sending a ConsensusNewHeightEvent to the p2p module
// This is necessary because the debug client is not a validator and has no consensus module but it has to update the peerstore
// depending on the changes in the validator set.
// TODO(#613): Make the debug client mimic a full node.
// TECHDEBT: This may no longer be required (https://github.com/pokt-network/pocket/pull/891/files#r1262710098)
func sendConsensusNewHeightEventToP2PModule(height uint64, bus modules.Bus) error {
	newHeightEvent, err := messaging.PackMessage(&messaging.ConsensusNewHeightEvent{Height: height})
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to pack consensus new height event")
	}
	return bus.GetP2PModule().HandleEvent(newHeightEvent.Content)
}
