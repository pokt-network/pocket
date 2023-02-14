package p2p

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
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

	case messaging.StateMachineTransitionEventType:
		stateMachineTransitionEvent, ok := evt.(*messaging.StateMachineTransitionEvent)
		if !ok {
			return fmt.Errorf("failed to cast event to StateMachineTransitionEvent")
		}

		if stateMachineTransitionEvent.NewState == string(coreTypes.StateMachineState_P2P_Bootstrapping) {
			addrBook := m.network.GetAddrBook()
			if len(addrBook) == 0 {
				m.logger.Warn().Msg("No peers in addrbook, bootstrapping")

				err := bootstrap(m)
				if err != nil {
					return err
				}
			}
			if !isSelfInAddrBook(m.address, addrBook) {
				m.logger.Warn().Msg("Self address not found in addresbook, advertising")
				// TODO: (link libp2p issue) advertise node to network, populate internal addressbook adding self as first peer
			}
			if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_P2P_IsBootstrapped); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unknown event type: %s", event.MessageName())
	}

	return nil
}
