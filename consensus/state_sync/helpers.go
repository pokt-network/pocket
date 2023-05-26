package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

// SendStateSyncMessage sends a state sync message after converting to any proto, to the given peer
func (m *stateSync) sendStateSyncMessage(msg *typesCons.StateSyncMessage, dst cryptoPocket.Address) error {
	if anyMsg, err := anypb.New(msg); err != nil {
		return err
	} else if err := m.GetBus().GetP2PModule().Send(dst, anyMsg); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrSendMessage.Error())
		return err
	}
	return nil
}

func (m *stateSync) getValidatorsAtHeight(height uint64) ([]*coreTypes.Actor, error) {
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(height))
	if err != nil {
		return nil, err
	}
	defer readCtx.Release()
	return readCtx.GetAllValidators(int64(height))
}

// TECHDEBT(#686): This should be an ongoing background passive state sync process but just
// capturing the available messages at the time that this function was called is good enough for now.
func (m *stateSync) getAggregatedStateSyncMetadata() typesCons.StateSyncMetadataResponse {
	chanLen := len(m.metadataReceived)
	m.logger.Info().Msgf("Looping over %d state sync metadata responses", chanLen)

	minHeight, maxHeight := uint64(1), uint64(1)
	for i := 0; i < chanLen; i++ {
		metadata := <-m.metadataReceived
		if metadata.MaxHeight > maxHeight {
			maxHeight = metadata.MaxHeight
		}
		if metadata.MinHeight < minHeight {
			minHeight = metadata.MinHeight
		}
	}

	return typesCons.StateSyncMetadataResponse{
		PeerAddress: "unused_aggregated_metadata_address",
		MinHeight:   minHeight,
		MaxHeight:   maxHeight,
	}
}
