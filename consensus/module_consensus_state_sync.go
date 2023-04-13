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

// TODO(#352): Implement this function, currently a placeholder.
// commitReceivedBlocks commits the blocks received from the blocksReceived channel
// it is intended to be run as a background process
func (m *consensusModule) blockApplicationLoop() {
	// runs as a background process in consensus module
	// listens on the blocksReceived channel
	// commits the received block
}

// TODO(#352): Implement this function, currently a placeholder.
// metadataSyncLoop periodically sends metadata requests to its peers
// it is intended to be run as a background process
func (m *consensusModule) metadataSyncLoop() {
	// runs as a background process in consensus module
	// requests metadata from peers
	// sends received metadata to the metadataReceived channel
}
