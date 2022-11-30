package shared

import (
	"log"

	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/rpc"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
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

func CreateNode(runtime modules.RuntimeMgr) (modules.Module, error) {
	return new(Node).Create(runtime)
}

func (m *Node) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	persistenceMod, err := persistence.Create(runtimeMgr)
	if err != nil {
		return nil, err
	}

	p2pMod, err := p2p.Create(runtimeMgr)
	if err != nil {
		return nil, err
	}

	utilityMod, err := utility.Create(runtimeMgr)
	if err != nil {
		return nil, err
	}

	consensusMod, err := consensus.Create(runtimeMgr)
	if err != nil {
		return nil, err
	}

	telemetryMod, err := telemetry.Create(runtimeMgr)
	if err != nil {
		return nil, err
	}

	loggerMod, err := logger.Create(runtimeMgr)
	if err != nil {
		return nil, err
	}

	rpcMod, err := rpc.Create(runtimeMgr)
	if err != nil {
		return nil, err
	}

	bus, err := CreateBus(
		runtimeMgr,
		persistenceMod.(modules.PersistenceModule),
		p2pMod.(modules.P2PModule),
		utilityMod.(modules.UtilityModule),
		consensusMod.(modules.ConsensusModule),
		telemetryMod.(modules.TelemetryModule),
		loggerMod.(modules.LoggerModule),
		rpcMod.(modules.RPCModule),
	)
	if err != nil {
		return nil, err
	}
	addr, err := p2pMod.(modules.P2PModule).GetAddress()
	if err != nil {
		return nil, err
	}
	return &Node{
		bus:        bus,
		p2pAddress: addr,
	}, nil
}

func (node *Node) Start() error {
	log.Println("About to start pocket node modules...")

	// IMPORTANT: Order of module startup here matters

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

	log.Println("About to start pocket node main loop...")

	// While loop lasting throughout the entire lifecycle of the node to handle asynchronous events
	for {
		event := node.GetBus().GetBusEvent()
		if err := node.handleEvent(event); err != nil {
			log.Println("Error handling event: ", err)
		}
	}
}

func (node *Node) Stop() error {
	log.Println("Stopping pocket node...")
	return nil
}

func (m *Node) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *Node) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (node *Node) handleEvent(message *messaging.PocketEnvelope) error {
	contentType := message.GetContentType()
	switch contentType {
	case messaging.NodeStartedEventType:
		log.Println("[NOOP] Received NodeStartedEvent")
	case consensus.HotstuffMessageContentType:
		return node.GetBus().GetConsensusModule().HandleMessage(message.Content)
	case messaging.DebugMessageEventType:
		return node.handleDebugMessage(message)
	default:
		log.Printf("[WARN] Unsupported message content type: %s \n", contentType)
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
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
		fallthrough
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE:
		fallthrough
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW:
		fallthrough
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE:
		return node.GetBus().GetConsensusModule().HandleDebugMessage(debugMessage)
	// Persistence Debug
	case messaging.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		return node.GetBus().GetPersistenceModule().HandleDebugMessage(debugMessage)
	// Default Debug
	default:
		log.Printf("Debug message: %s \n", debugMessage.Message)
	}

	return nil
}

func (node *Node) GetModuleName() string {
	return mainModuleName
}

func (node *Node) GetP2PAddress() cryptoPocket.Address {
	return node.p2pAddress
}
