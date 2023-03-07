package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Helper function for broadcasting state sync messages to the all peers known to the node
// It is used for:
//
//		requesting for metadata, via the periodicSynch() function
//	 	requesting for blocks, via the StartSynching() function
func (m *stateSync) broadCastStateSyncMessage(stateSyncMsg *typesCons.StateSyncMessage, height uint64) error {
	m.logger.Info().Fields(
		map[string]any{
			"height": height,
			"nodeId": m.GetBus().GetConsensusModule().GetNodeId(),
		},
	).Msg("📣 Broadcasting state sync message GOKHANSA📣")

	anyMessage, err := codec.GetCodec().ToAny(stateSyncMsg)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateConsensusMessage.Error())
		return err
	}

	validators, err := m.GetBus().GetConsensusModule().GetValidatorsAtHeight(height)
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

	fields := map[string]any{
		"height":     height,
		"peerId":     peerId,
		"proto_type": getMessageType(stateSyncMsg),
	}

	m.logger.Info().Fields(fields).Msg("Sending StateSync Message")
	return m.sendToPeer(anyMsg, peerId)
}

// Helper function for sending state sync messages
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

// func (m *stateSync) validateBlock(block *types.Block) error {

// 	//stateHash := block.BlockHeader.StateHash
// 	//quorumCertificate := block.BlockHeader.QuorumCertificate

// 	// validate quorum certificate

// 	// validate block

// 	// apply block

// 	return nil
// }
