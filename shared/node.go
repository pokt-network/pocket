package shared

import (
	"context"
	"time"

	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/ibc"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/rpc"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/state_machine"
	"github.com/pokt-network/pocket/telemetry"
	"github.com/pokt-network/pocket/utility"
	"go.uber.org/multierr"
)

const (
	mainModuleName       = "main"
	eventHandlingTimeout = 30 * time.Second // see usage for a description
)

type Node struct {
	bus        modules.Bus
	p2pAddress cryptoPocket.Address
}

func NewNodeWithP2PAddress(address cryptoPocket.Address) *Node {
	return &Node{p2pAddress: address}
}

func CreateNode(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(Node).Create(bus, options...)
}

func (m *Node) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	for _, mod := range []func(modules.Bus, ...modules.ModuleOption) (modules.Module, error){
		state_machine.Create,
		persistence.Create,
		utility.Create,
		consensus.Create,
		telemetry.Create,
		logger.Create,
		rpc.Create,
		p2p.Create,
		ibc.Create,
	} {
		if _, err := mod(bus); err != nil {
			return nil, err
		}
	}

	addr, err := bus.GetP2PModule().GetAddress()
	if err != nil {
		return nil, err
	}

	return &Node{
		bus:        bus,
		p2pAddress: addr,
	}, nil
}

func (node *Node) Start() error {
	logger.Global.Info().Msg("About to start pocket node modules...")

	// IMPORTANT: Order of module startup here matters

	if err := node.GetBus().GetStateMachineModule().Start(); err != nil {
		return err
	}

	if err := node.GetBus().GetTelemetryModule().Start(); err != nil {
		return err
	}

	if err := node.GetBus().GetPersistenceModule().Start(); err != nil {
		return err
	}

	if err := node.GetBus().GetP2PModule().Start(); err != nil {
		return err
	}

	if err := node.GetBus().GetUtilityModule().Start(); err != nil {
		return err
	}

	if err := node.GetBus().GetConsensusModule().Start(); err != nil {
		return err
	}

	if err := node.GetBus().GetRPCModule().Start(); err != nil {
		return err
	}

	if err := node.GetBus().GetIBCModule().Start(); err != nil {
		return err
	}

	// The first event signaling that the node has started
	signalNodeStartedEvent, err := messaging.PackMessage(&messaging.NodeStartedEvent{})
	if err != nil {
		return err
	}
	node.GetBus().PublishEventToBus(signalNodeStartedEvent)

	logger.Global.Info().Msg("About to start pocket node main loop...")

	// A while loop lasting throughout the entire lifecycle of the node to handle asynchronous events
	// send between modules or external participants.
	for {
		//
		event := node.GetBus().GetBusEvent()
		clock := node.GetBus().GetRuntimeMgr().GetClock()
		ctx, cancel := clock.WithTimeout(context.TODO(), eventHandlingTimeout)

		// `node.handleEvent`` is a blocking call, and the entrypoint into all the operations inside the node.
		// It is run in a goroutine to allow setting a deadline for message handling and get visibility into
		// bugs/issue including deadlocks and other concurrency issues.
		go func() {
			if err := node.handleEvent(event); err != nil {
				logger.Global.Error().Err(err).Msg("Error handling event")
			}
			cancel()
		}()

		// Block the node event handler from continuing until the event has been handled or the deadline has been reached.
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				logger.Global.Error().Msgf("Event handling timed out: %v", event)
				cancel()
			}
		case <-clock.After(eventHandlingTimeout + 1*time.Second):
			cancel()
		}
	}
}

func (node *Node) Stop() error {
	logger.Global.Info().Msg("Stopping pocket node...")
	return nil
}

func (m *Node) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *Node) GetBus() modules.Bus {
	if m.bus == nil {
		logger.Global.Fatal().Msg("PocketBus is not initialized")
	}
	return m.bus
}

// TECHDEBT: The `shared` package has dependencies on types in the individual modules.
// TODO: Move all message types this is dependant on to the `messaging` package
func (node *Node) handleEvent(message *messaging.PocketEnvelope) error {
	contentType := message.GetContentType()
	logger.Global.Debug().Fields(map[string]any{
		"message":     message,
		"contentType": contentType,
	}).Msg("node handling event")

	switch contentType {
	case messaging.NodeStartedEventType:
		logger.Global.Info().Msg("Received NodeStartedEvent")
		if err := node.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Start); err != nil {
			return err
		}
	case messaging.HotstuffMessageContentType:
		return node.GetBus().GetConsensusModule().HandleMessage(message.Content)
	case messaging.StateSyncMessageContentType:
		return node.GetBus().GetConsensusModule().HandleStateSyncMessage(message.Content)
	case messaging.TxGossipMessageContentType:
		return node.GetBus().GetUtilityModule().HandleUtilityMessage(message.Content)
	case messaging.DebugMessageEventType:
		if err := node.GetBus().GetP2PModule().HandleEvent(message.Content); err != nil {
			return err
		}
		if err := node.handleDebugMessage(message); err != nil {
			return err
		}
	case messaging.ConsensusNewHeightEventType:
		err_p2p := node.GetBus().GetP2PModule().HandleEvent(message.Content)
		err_ibc := node.GetBus().GetIBCModule().HandleEvent(message.Content)
		// TODO: Remove this lib once we move to Go 1.2
		return multierr.Combine(err_p2p, err_ibc)
	case messaging.StateMachineTransitionEventType:
		err_consensus := node.GetBus().GetConsensusModule().HandleEvent(message.Content)
		err_p2p := node.GetBus().GetP2PModule().HandleEvent(message.Content)
		// TODO: Remove this lib once we move to Go 1.2
		return multierr.Combine(err_consensus, err_p2p)
	default:
		logger.Global.Warn().Msgf("Unsupported message content type: %s", contentType)
	}
	return nil
}

func (node *Node) handleDebugMessage(message *messaging.PocketEnvelope) error {
	// Consensus Debug
	debugMessage, err := messaging.UnpackMessage[*messaging.DebugMessage](message)
	if err != nil {
		return err
	}
	switch debugMessage.Action {
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS,
		messaging.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE,
		messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
		messaging.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE,
		messaging.DebugMessageAction_DEBUG_CONSENSUS_SEND_BLOCK_REQ,
		messaging.DebugMessageAction_DEBUG_CONSENSUS_SEND_METADATA_REQ:
		return node.GetBus().GetConsensusModule().HandleDebugMessage(debugMessage)
	// Persistence Debug
	case messaging.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		return node.GetBus().GetPersistenceModule().HandleDebugMessage(debugMessage)
	// Default Debug
	default:
		logger.Global.Debug().Msgf("Received DebugMessage: %s", debugMessage.Message)
	}

	return nil
}

func (node *Node) GetModuleName() string {
	return mainModuleName
}

func (node *Node) GetP2PAddress() cryptoPocket.Address {
	return node.p2pAddress
}
