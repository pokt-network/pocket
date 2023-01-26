package state_sync

import (
	"encoding/binary"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Helper function for sending state sync message to
func (m *stateSync) sendToPeer(msg *anypb.Any, peerId cryptoPocket.Address) error {
	if err := m.GetBus().GetP2PModule().Send(peerId, msg); err != nil {
		m.nodeLogError(typesCons.ErrSendMessage.Error(), err)
		return err
	}
	return nil
}

// This is copy paste from Peristantace/state_test.go
// TODO Check if can be done without copy paste
func heightToBytes(height uint64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, height)
	return heightBytes
}

// func heightFromBytes(heightBz []byte) uint64 {
// 	return binary.LittleEndian.Uint64(heightBz)
// 	//return new(big.Int).SetBytes(heightBz).Uint64()
// }
