package p2p

import (
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
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

		if isStaked, err := m.isStakedActor(); err != nil {
			return err
		} else if !isStaked {
			return nil // unstaked actors do not use RainTree and therefore do not need to update this router
		}

		oldPeerList := m.stakedActorRouter.GetPeerstore().GetPeerList()
		pstoreProvider, err := m.getPeerstoreProvider()
		if err != nil {
			return err
		}

		updatedPeerstore, err := pstoreProvider.GetStakedPeerstoreAtHeight(consensusNewHeightEvent.Height)
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
			staked, err := m.isStakedActor()
			if err != nil {
				return err
			}
			if staked {
				// TECHDEBT(#859): this will never happen as the peerstore is
				// seeded from consensus during P2P module construction.
				if m.stakedActorRouter.GetPeerstore().Size() == 0 {
					m.logger.Warn().Msg("No peers in peerstore, bootstrapping")

					if err := m.bootstrap(); err != nil {
						return err
					}
				}
			}

			// TECHDEBT(#859): for unstaked actors, unstaked actor (background)
			// router bootstrapping SHOULD complete before the event below is sent.

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
