package state_sync

import (
	"encoding/binary"
	"errors"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *stateSyncModule) sendToPeer(msg *anypb.Any, peerId string) error {
	//Seperation between nodeId and peerId must be clear

	// TODO: Check if this is needed (added since it was added in consensus module sendToNode function)
	if !m.GetBus().GetConsensusModule().IsLeaderSet() {
		m.nodeLogError(typesCons.ErrNilLeaderId.Error(), nil)
		return errors.New(typesCons.ErrNilLeaderId.Error())
	}

	//nodeId := typesCons.NodeId(m.bus.GetConsensusModule().GetNodeIdFromNodeAddress(peerId))
	nodeIdint, err := m.bus.GetConsensusModule().GetNodeIdFromNodeAddress(peerId)
	if err != nil {
		return err
	}

	nodeId := typesCons.NodeId(nodeIdint)

	m.nodeLog(typesCons.SendingStateSyncMessage(msg, nodeId))
	anyConsensusMessage, err := codec.GetCodec().ToAny(msg)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateConsensusMessage.Error(), err)
		return err
	}
	if err := m.GetBus().GetP2PModule().Send(cryptoPocket.AddressFromString(peerId), anyConsensusMessage); err != nil {
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
