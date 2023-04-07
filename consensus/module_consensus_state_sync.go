package consensus

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.ConsensusStateSync = &consensusModule{}

func (m *consensusModule) GetNodeIdFromNodeAddress(peerId string) (uint64, error) {
	validators, err := m.GetValidatorsAtHeight(m.CurrentHeight())
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

// // IsSynced implements the interface function for checking if the node is synced with the network.
// func (m *consensusModule) IsSynced() (bool, error) {
// 	currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()
// 	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight - 1)) // Unknown height
// 	if err != nil {
// 		return false, err
// 	}
// 	defer readCtx.Release()

// 	maxPersistedHeight, err := readCtx.GetMaximumBlockHeight()
// 	if err != nil {
// 		return false, err
// 	}

// 	maxSeenHeight := m.GetAggregatedMetadata().MaxHeight

// 	return maxPersistedHeight == maxSeenHeight, nil
// }

func (m *consensusModule) GetAggregatedStateSyncMetadata() (minHeight, maxHeight uint64) {
	minHeight, maxHeight = 1, 1

	chanLen := len(m.MetadataReceived)

	for i := 0; i < chanLen; i++ {
		metadata := <-m.MetadataReceived
		if metadata.MaxHeight > maxHeight {
			maxHeight = metadata.MaxHeight
		}
		if metadata.MinHeight < minHeight {
			minHeight = metadata.MinHeight
		}
	}

	return minHeight, maxHeight
}
