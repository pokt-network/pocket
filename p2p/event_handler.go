package p2p

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *p2pModule) HandleEvent(event *anypb.Any) error {
	evt, err := codec.GetCodec().FromAny(event)
	if err != nil {
		return err
	}

	switch event.MessageName() {
	case messaging.ConsensusNewHeightEventType:
		consensusNewHeightEvent, ok := evt.(*messaging.ConsensusNewHeightEvent)
		if !ok {
			return fmt.Errorf("failed to cast event to ConsensusNewHeightEvent")
		}

		addrBook := m.network.GetAddrBook()
		newAddrBook, err := m.addrBookProvider.GetStakedAddrBookAtHeight(consensusNewHeightEvent.Height)

		if err != nil {
			return err
		}

		added, removed := getAddrBookDelta(addrBook, newAddrBook)
		for _, add := range added {
			if err := m.network.AddPeerToAddrBook(add); err != nil {
				return err
			}
		}
		for _, rm := range removed {
			if err := m.network.RemovePeerFromAddrBook(rm); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unknown event type: %s", event.MessageName())
	}

	return nil
}
