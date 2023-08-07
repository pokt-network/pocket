package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
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

// TECHDEBT(#686): This should be an ongoing background passive state sync process.
// For now, aggregating the messages when requests is good enough.
func (m *stateSync) getAggregatedStateSyncMetadata() (minHeight, maxHeight uint64) {
	chanLen := len(m.metadataReceived)
	m.logger.Info().
		Int16("num_state_sync_metadata_messages", int16(chanLen)).
		Msgf("About to loop overstate sync metadata responses")

	for i := 0; i < chanLen; i++ {
		metadata := <-m.metadataReceived
		if metadata.MaxHeight > maxHeight {
			maxHeight = metadata.MaxHeight
		}
		if metadata.MinHeight < minHeight {
			minHeight = metadata.MinHeight
		}
	}
	m.logger.Info().Fields(map[string]any{
		"min_height": minHeight,
		"max_height": maxHeight,
	}).Msg("Finished aggregating state sync metadata")
	return
}
