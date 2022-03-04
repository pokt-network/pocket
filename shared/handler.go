package shared

import (
	"log"

	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func (node *Node) handleEvent(event *types.PocketEvent) error {
	switch event.Topic {
	case types.PocketTopic_CONSENSUS_MESSAGE_TOPIC:
		return node.GetBus().GetConsensusModule().HandleMessage(event.Data)
	case types.PocketTopic_DEBUG_TOPIC:
		return node.handleDebugEvent(event.Data)
	default:
		log.Printf("[WARN] Unsupported PocketEvent topic: %s \n", event.Topic)
	}
	return nil
}

func (node *Node) handleDebugEvent(anyMessage *anypb.Any) error {
	var debugMessage types.DebugMessage
	err := anypb.UnmarshalTo(anyMessage, &debugMessage, proto.UnmarshalOptions{})
	if err != nil {
		return err
	}

	switch debugMessage.Action {
	case types.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
		fallthrough
	case types.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE:
		fallthrough
	case types.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW:
		fallthrough
	case types.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE:
		return node.GetBus().GetConsensusModule().HandleDebugMessage(&debugMessage)
	default:
		log.Printf("Debug message: %s \n", debugMessage.Message)
	}

	return nil
}
