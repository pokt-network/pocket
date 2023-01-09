package state_sync

import (
	"errors"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

/*
type NodeId uint64

type ValAddrToIdMap map[string]NodeId // Mapping from hex encoded address to an integer node id.
type IdToValAddrMap map[NodeId]string // Mapping from node id to a hex encoded string address.

I need to have this: idToValAddrMap

 I believe that PeerId is equivalent to a node's cryptographic address. In #413 code review I should have probably signaled the fact that the naming can be confusing.

*/

func (m *stateSyncModule) sendToPeer(msg *anypb.Any, peerId string) error {
	//Seoeration between nodeId and peerId must be clear

	// TODO: Check if this is needed (added since it was added in consensus module sendToNode function)
	if !m.GetBus().GetConsensusModule().IsLeaderSet() {
		m.nodeLogError(typesCons.ErrNilLeaderId.Error(), nil)
		return errors.New(typesCons.ErrNilLeaderId.Error())
	}

	nodeId := typesCons.NodeId(m.bus.GetConsensusModule().GetNodeIdFromNodeAddress(peerId))

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
