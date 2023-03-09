package shared

import (
	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/libp2p"
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
)

const (
	mainModuleName = "main"
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
	// TECHDEBT: simplify after P2P module consolidation.
	useLibP2P := bus.GetRuntimeMgr().GetConfig().UseLibP2P
	p2pCreate := p2p.Create
	if useLibP2P {
		p2pCreate = libp2p.Create
	}

	for _, mod := range []func(modules.Bus, ...modules.ModuleOption) (modules.Module, error){
		state_machine.Create,
		persistence.Create,
		utility.Create,
		consensus.Create,
		telemetry.Create,
		logger.Create,
		rpc.Create,
		p2pCreate,
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

	// The first event signaling that the node has started
	signalNodeStartedEvent, err := messaging.PackMessage(&messaging.NodeStartedEvent{})
	if err != nil {
		return err
	}
	node.GetBus().PublishEventToBus(signalNodeStartedEvent)

	logger.Global.Info().Msg("About to start pocket node main loop...")

	// While loop lasting throughout the entire lifecycle of the node to handle asynchronous events
	for {
		event := node.GetBus().GetBusEvent()
		if err := node.handleEvent(event); err != nil {
			logger.Global.Error().Err(err).Msg("Error handling event")
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
	switch contentType {
	case messaging.NodeStartedEventType:
		logger.Global.Info().Msg("Received NodeStartedEvent")
		if err := node.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Start); err != nil {
			return err
		}
	case consensus.HotstuffMessageContentType:
		return node.GetBus().GetConsensusModule().HandleMessage(message.Content)
	case consensus.StateSyncMessageContentType:
		return node.GetBus().GetConsensusModule().HandleStateSyncMessage(message.Content)
	case messaging.TxGossipMessageContentType:
		return node.GetBus().GetUtilityModule().HandleUtilityMessage(message.Content)
	case messaging.DebugMessageEventType:
		return node.handleDebugMessage(message)
	case messaging.ConsensusNewHeightEventType, messaging.StateMachineTransitionEventType:
		return node.GetBus().GetP2PModule().HandleEvent(message.Content)
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
