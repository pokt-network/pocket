package consensus

import (
	"fmt"
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"google.golang.org/protobuf/types/known/anypb"
)

// Implementations of the type ConsensusPacemaker interface
// SetHeight, SeetRound, SetStep are implemented for ConsensusDebugModule
func (m *consensusModule) ClearLeaderMessagesPool() {
	m.clearLeader()
	m.clearMessagesPool()
}

func (m *consensusModule) ResetForNewHeight() {
	m.round = 0
	m.block = nil
	m.prepareQC = nil
	m.lockedQC = nil
}

func (m *consensusModule) ReleaseUtilityContext() error {
	if m.utilityContext != nil {
		if err := m.utilityContext.Release(); err != nil {
			log.Println("[WARN] Failed to release utility context: ", err)
			return err
		}
		m.utilityContext = nil
	}

	return nil
}

func (m *consensusModule) BroadcastMessageToNodes(msg *anypb.Any) error {
	msgCodec, err := codec.GetCodec().FromAny(msg)
	if err != nil {
		return err
	}

	broadcastMessage, ok := msgCodec.(*typesCons.HotstuffMessage)
	if !ok {
		return fmt.Errorf("failed to cast message to HotstuffMessage")
	}
	m.broadcastToValidators(broadcastMessage)

	return nil
}

func (m *consensusModule) IsLeader() bool {
	return m.leaderId != nil && *m.leaderId == m.nodeId
}

func (m *consensusModule) IsLeaderSet() bool {
	return m.leaderId != nil
}

func (m *consensusModule) NewLeader(msg *anypb.Any) error {
	msgCodec, err := codec.GetCodec().FromAny(msg)
	if err != nil {
		return err
	}

	message, ok := msgCodec.(*typesCons.HotstuffMessage)
	if !ok {
		return fmt.Errorf("failed to cast message to HotstuffMessage")
	}

	return m.electNextLeader(message)
}

func (m *consensusModule) GetPrepareQC() *anypb.Any {
	anyProto, err := anypb.New(m.prepareQC)
	if err != nil {
		log.Println("[WARN] NewHeight: Failed to convert paceMaker message to proto: ", err)
		return nil
	}
	return anyProto

}

func (m *consensusModule) GetNodeId() uint64 {
	return uint64(m.nodeId)
}
