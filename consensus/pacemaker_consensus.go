package consensus

import (
	"fmt"
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"google.golang.org/protobuf/types/known/anypb"
)

// Implementations of the type ConsensusPacemaker interface
// SetHeight, SetRound, SetStep are implemented for ConsensusDebugModule
func (m *consensusModule) ResetRound() {
	m.clearLeader()
	m.clearMessagesPool()
}

// This function resets the current state of the consensus module, called by pacemaker submodule before node proceeds to the next view.
func (m *consensusModule) ResetForNewHeight() {
	m.round = 0
	m.block = nil
	m.prepareQC = nil
	m.lockedQC = nil
}

// This function releases consensus module's utility context, called by pacemaker module
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

func (m *consensusModule) BroadcastMessageToValidators(msg *anypb.Any) error {
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

func (m *consensusModule) IsPrepareQCNil() bool {
	return m.prepareQC == nil
}

func (m *consensusModule) GetPrepareQC() (*anypb.Any, error) {
	anyProto, err := anypb.New(m.prepareQC)
	if err != nil {
		return nil, fmt.Errorf("[WARN] NewHeight: Failed to convert paceMaker message to proto: %s", err)
	}
	return anyProto, nil
}

func (m *consensusModule) GetNodeId() uint64 {
	return uint64(m.nodeId)
}
