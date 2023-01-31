package state_sync

import (
	"encoding/binary"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Helper function for sending state sync messages
func (m *stateSync) sendToPeer(msg *anypb.Any, peerId cryptoPocket.Address) error {
	if err := m.GetBus().GetP2PModule().Send(peerId, msg); err != nil {
		m.nodeLogError(typesCons.ErrSendMessage.Error(), err)
		return err
	}
	return nil
}

// TODO Check if heightToBytes can be a unified common function, as there is identical function in peristantace/state_test.go and persistence/block.go:
func heightToBytes(height uint64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, height)
	return heightBytes
}
