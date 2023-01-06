package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *consensusModule) HandleStateSyncMessage(stateSyncMessageAny *anypb.Any) error {
	m.m.Lock()
	defer m.m.Unlock()

	switch stateSyncMessageAny.MessageName() {
	case StateSyncMessageContentType:

		msg, err := codec.GetCodec().FromAny(stateSyncMessageAny)
		if err != nil {
			return err
		}

		stateSyncMessage, ok := msg.(*typesCons.StateSyncMessage)
		if !ok {
			return fmt.Errorf("failed to cast message to HotstuffMessage")
		}

		if err := m.handleStateSyncMessage(stateSyncMessage); err != nil {
			return err
		}

	default:
		return typesCons.ErrUnknownStateSyncMessageType(stateSyncMessageAny.MessageName())
	}

	return nil
}

func (m *consensusModule) handleStateSyncMessage(stateSyncMessage *typesCons.StateSyncMessage) error {

	switch stateSyncMessage.MsgType {
	case typesCons.StateSyncMessageType_STATE_SYNC_METADATA_REQUEST:
		msg := stateSyncMessage.GetMetadataReq()
		err := m.stateSync.HandleStateSyncMetadataRequest(msg)
		if err != nil {
			return err
		}
	case typesCons.StateSyncMessageType_STATE_SYNC_METADATA_RESPONSE:
		msg := stateSyncMessage.GetMetadataRes()
		err := m.stateSync.HandleStateSyncMetadataResponse(msg)
		if err != nil {
			return err
		}
	case typesCons.StateSyncMessageType_STATE_SYNC_GET_BLOCK_REQUEST:
		msg := stateSyncMessage.GetGetBlockReq()
		err := m.stateSync.HandleGetBlockRequest(msg)
		if err != nil {
			return err
		}
	case typesCons.StateSyncMessageType_STATE_SYNC_GET_BLOCK_RESPONSE:
		msg := stateSyncMessage.GetGetBlockRes()
		err := m.stateSync.HandleGetBlockResponse(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
func (m *consensusModule) handlePaceMakerMessage(pacemakerMsg *typesCons.PacemakerMessage) error {
	log.Print("\n Pacemaker Event Received! \n")

	switch pacemakerMsg.Action {
	// case typesCons.PacemakerMessageType_PACEMAKER_MESSAGE_SET_HEIGHT:
	// 	m.height = pacemakerMsg.GetHeight().Height
	// 	log.Printf("handlePaceMakerMessage Height is: %d", m.height)
	// case typesCons.PacemakerMessageType_PACEMAKER_MESSAGE_SET_ROUND:
	// 	m.round = pacemakerMsg.GetRound().Round
	// case typesCons.PacemakerMessageType_PACEMAKER_MESSAGE_SET_STEP:
	// 	m.step = typesCons.HotstuffStep(pacemakerMsg.GetStep().Step)
	// case typesCons.PacemakerMessageType_PACEMAKER_MESSAGE_RESET_FOR_NEW_HEIGHT:
	// 	m.resetForNewHeight()
	// case typesCons.PacemakerMessageType_PACEMAKER_MESSAGE_CLEAR_LEADER_MESSAGE_POOL:
	// 	m.clearLeader()
	// 	m.clearMessagesPool()
	case typesCons.PacemakerMessageType_PACEMAKER_MESSAGE_RELEASE_UTILITY_CONTEXT:
		if m.utilityContext != nil {
			if err := m.utilityContext.Release(); err != nil {
				log.Println("[WARN] Failed to release utility context: ", err)
				return err
			}
			m.utilityContext = nil
		}
	case typesCons.PacemakerMessageType_PACEMAKER_MESSAGE_BROADCAST_HOTSTUFF_MESSAGE_TO_NODES:
		m.broadcastToNodes(pacemakerMsg.GetMessage())
	default:
		log.Printf("\n\n Unexpected case, message Action: %s \n\n", &pacemakerMsg.Action)
	}
	return nil
}
*/
