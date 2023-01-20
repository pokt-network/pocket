package state_sync

import (
	"encoding/binary"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *stateSync) sendToPeer(msg *anypb.Any, peerId cryptoPocket.Address) error {
	if err := m.GetBus().GetP2PModule().Send(peerId, msg); err != nil {
		m.nodeLogError(typesCons.ErrSendMessage.Error(), err)
		return err
	}

	return nil
}

// This is copy paste from Peristantace/state_test.go
// TODO Check if can be done without copy paste
func heightToBytes(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
	return heightBytes
}
