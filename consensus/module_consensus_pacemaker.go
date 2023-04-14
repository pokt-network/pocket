package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.ConsensusPacemaker = &consensusModule{}

// ResetRound resets the hotstuff view / round at either the current or new height depending on the var prvodied
func (m *consensusModule) ResetRound(isNewHeight bool) {
	m.leaderId = nil
	m.clearMessagesPool()
	m.step = 0
	if isNewHeight {
		m.round = 0
		m.block = nil
		m.prepareQC = nil
		m.lockedQC = nil
	}
}

// This function releases consensus module's utility unitOfWork, called by pacemaker module
func (m *consensusModule) ReleaseUtilityUnitOfWork() error {
	utilityUnitOfWork := m.utilityUnitOfWork
	if utilityUnitOfWork == nil {
		m.logger.Debug().Msg("Try to release a nil utilityUnitOfWork... Ideally this should not happen")
		return nil
	}
	if err := utilityUnitOfWork.Release(); err != nil {
		return err
	}
	m.utilityUnitOfWork = nil
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
		return nil, fmt.Errorf("failed to convert paceMaker message to proto: %w", err)
	}
	return anyProto, nil
}

func (m *consensusModule) GetNodeId() uint64 {
	return uint64(m.nodeId)
}
