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

// TODO (#352): Implement this function, currently a placeholder.
func (m *consensusModule) commitReceivedBlocks() {
	// runs in the backgroun of consensus module
	// listens on the blocksReceived channel
	// commits the block
	// sends the committed block information to the to sync's channel
}

// TODO (#352): Implement this function, currently a placeholder.
// metadataSyncLoop periodically queries the network to see if it is behind
func (m *consensusModule) metadataSyncLoop() {
	// runs in the background of consensus module
	// requests metadata from peers
	// sends received metadata to the metadataReceived channel
}
