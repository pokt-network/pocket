package p2p

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/types/known/anypb"
)

// CONSIDERATION(#576): making this part of some new `ConnManager`.
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

		oldPeerList := m.stakedActorRouter.GetPeerstore().GetPeerList()
		updatedPeerstore, err := m.pstoreProvider.GetStakedPeerstoreAtHeight(consensusNewHeightEvent.Height)
		if err != nil {
			return err
		}

		added, removed := oldPeerList.Delta(updatedPeerstore.GetPeerList())
		for _, add := range added {
			if err := m.stakedActorRouter.AddPeer(add); err != nil {
				return err
			}
		}
		for _, rm := range removed {
			if err := m.stakedActorRouter.RemovePeer(rm); err != nil {
				return err
			}
		}

	case messaging.StateMachineTransitionEventType:
		stateMachineTransitionEvent, ok := evt.(*messaging.StateMachineTransitionEvent)
		if !ok {
			return fmt.Errorf("failed to cast event to StateMachineTransitionEvent")
		}

		m.logger.Debug().Fields(messaging.TransitionEventToMap(stateMachineTransitionEvent)).Msg("Received state machine transition event")

		if stateMachineTransitionEvent.NewState == string(coreTypes.StateMachineState_P2P_Bootstrapping) {
			if m.stakedActorRouter.GetPeerstore().Size() == 0 {
				m.logger.Warn().Msg("No peers in addrbook, bootstrapping")

				if err := m.bootstrap(); err != nil {
					return err
				}
			}
			m.logger.Info().Bool("TODO", true).Msg("Advertise self to network")
			if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_P2P_IsBootstrapped); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unknown event type: %s", event.MessageName())
	}

	return nil
}
