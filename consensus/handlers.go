package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *consensusModule) HandleMessage(message *anypb.Any) error {
	var consensusMessage types_consensus.ConsensusMessage
	err := anypb.UnmarshalTo(message, &consensusMessage, proto.UnmarshalOptions{})
	if err != nil {
		return err
	}

	switch consensusMessage.Type {
	case HotstuffMessage:
		var hotstuffMessage types_consensus.HotstuffMessage
		err := anypb.UnmarshalTo(consensusMessage.Message, &hotstuffMessage, proto.UnmarshalOptions{})
		if err != nil {
			return err
		}
		m.handleHotstuffMessage(&hotstuffMessage)
	case UtilityMessage:
		m.handleTransaction(consensusMessage.Message)
	default:
		return fmt.Errorf("unknown consensus message type: %v", consensusMessage.Type)
	}
	return nil
}
