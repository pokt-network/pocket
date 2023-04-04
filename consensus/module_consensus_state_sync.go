package consensus

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.ConsensusStateSync = &consensusModule{}

func (m *consensusModule) GetNodeIdFromNodeAddress(peerId string) (uint64, error) {
	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		// REFACTOR(#434): As per issue #434, once the new id is sorted out, this return statement must be changed
		return 0, err
	}

	valAddrToIdMap := typesCons.NewActorMapper(validators).GetValAddrToIdMap()
	return uint64(valAddrToIdMap[peerId]), nil
}

func (m *consensusModule) GetNodeAddress() string {
	return m.nodeAddress
}

// IsSynced implements the interface function for checking if the node is synced with the network.
func (m *consensusModule) IsSynced() (bool, error) {
	maxPersistedHeight, err := m.maxPersistedBlockHeight()
	if err != nil {
		return false, err
	}

	maxSeenHeight := m.stateSync.GetAggregatedStateSyncMetadata().MaxHeight

	return maxPersistedHeight == maxSeenHeight, nil
}
