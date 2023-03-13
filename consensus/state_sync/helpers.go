package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

// TODO (#352): Implement this function, currently a placeholder.
// Helper function for broadcasting state sync messages to the all peers known to the node:
//
//		requests for metadata using the `periodicMetadataSynch()` function
//	 	requests for blocks using the `StartSynching()` function
func (m *stateSync) broadcastStateSyncMessage(stateSyncMsg *typesCons.StateSyncMessage, height uint64) error {
	// TODO (#571): update with logger helper function
	m.logger.Info().Fields(
		map[string]any{
			"height": height,
			"nodeId": m.GetBus().GetConsensusModule().GetNodeId(),
		},
	).Msg("📣 Broadcasting state sync message... 📣")

	anyMessage, err := codec.GetCodec().ToAny(stateSyncMsg)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateConsensusMessage.Error())
		return err
	}

	validators, err := m.getValidatorsAtHeight(height)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrPersistenceGetAllValidators.Error())
	}

	// for _, val := range validators {
	// 	m.logger.Debug().Msgf("VAL: %s", val.Address)
	// 	if err := m.SendStateSyncMessage(stateSyncMsg, cryptoPocket.Address(val.Address), height); err != nil {
	// 		m.logger.Error().Err(err).Msg(typesCons.ErrSendMessage.Error())
	// 		return err
	// 	}
	// }

	for _, val := range validators {
		m.logger.Info().Fields(
			map[string]any{
				"val": val.GetAddress(),
			},
		).Msg("📣 Sneding state sync message 📣")
		if err := m.GetBus().GetP2PModule().Send(cryptoPocket.AddressFromString(val.GetAddress()), anyMessage); err != nil {
			m.logger.Error().Err(err).Msg(typesCons.ErrBroadcastMessage.Error())
		}
	}

	return nil
}

func (m *stateSync) SendStateSyncMessage(stateSyncMsg *typesCons.StateSyncMessage, peerId cryptoPocket.Address, height uint64) error {
	anyMsg, err := anypb.New(stateSyncMsg)
	if err != nil {
		return err
	}

	// TODO (#571): update with logger helper function
	fields := map[string]any{
		"height":     height,
		"peerId":     peerId,
		"proto_type": getMessageType(stateSyncMsg),
	}

	m.logger.Info().Fields(fields).Msg("Sending StateSync Message")
	return m.sendToPeer(anyMsg, peerId)
}

// Helper function for messages to the peers
func (m *stateSync) sendToPeer(msg *anypb.Any, peerId cryptoPocket.Address) error {
	if err := m.GetBus().GetP2PModule().Send(peerId, msg); err != nil {
		m.logger.Error().Msgf(typesCons.ErrSendMessage.Error(), err)
		return err
	}
	return nil
}

func getMessageType(msg *typesCons.StateSyncMessage) string {
	switch msg.Message.(type) {
	case *typesCons.StateSyncMessage_MetadataReq:
		return "StateSyncMetadataRequest"
	case *typesCons.StateSyncMessage_MetadataRes:
		return "StateSyncMetadataResponse"
	case *typesCons.StateSyncMessage_GetBlockReq:
		return "GetBlockRequest"
	case *typesCons.StateSyncMessage_GetBlockRes:
		return "GetBlockResponse"
	default:
		return "Unknown"
	}
}

func (m *stateSync) getValidatorsAtHeight(height uint64) ([]*coreTypes.Actor, error) {
	persistenceReadContext, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(height))
	if err != nil {
		return nil, err
	}
	defer persistenceReadContext.Close()

	return persistenceReadContext.GetAllValidators(int64(height))
}

func (m *stateSync) maximumPersistedBlockHeight() (uint64, error) {
	currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	persistenceContext, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return 0, err
	}
	defer persistenceContext.Close()

	maxHeight, err := persistenceContext.GetMaximumBlockHeight()
	if err != nil {
		return 0, err
	}

	return maxHeight, nil
}

func (m *stateSync) minimumPersistedBlockHeight() (uint64, error) {
	currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	persistenceContext, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return 0, err
	}
	defer persistenceContext.Close()

	maxHeight, err := persistenceContext.GetMinimumBlockHeight()
	if err != nil {
		return 0, err
	}

	return maxHeight, nil
}
