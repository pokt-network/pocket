package shared

import (
	"log"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/runtime"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/debug"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/telemetry"
	"github.com/pokt-network/pocket/utility"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	MainModuleName = "main"
)

type Node struct {
	bus        modules.Bus
	p2pAddress cryptoPocket.Address
}

func NewNodeWithP2PAddress(address cryptoPocket.Address) *Node {
	return &Node{p2pAddress: address}
}

func Create(configPath, genesisPath string, clock clock.Clock) (modules.Module, error) {
	return new(Node).Create(runtime.NewManagerFromFiles(configPath, genesisPath))
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

	bus, err := CreateBus(
		runtimeMgr,
		persistenceMod.(modules.PersistenceModule),
		p2pMod.(modules.P2PModule),
		utilityMod.(modules.UtilityModule),
		consensusMod.(modules.ConsensusModule),
		telemetryMod.(modules.TelemetryModule),
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

	// The first event signaling that the node has started
	signalNodeStartedEvent := &debug.PocketEvent{Topic: debug.PocketTopic_POCKET_NODE_TOPIC, Data: nil}
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

func (node *Node) handleEvent(event *debug.PocketEvent) error {
	switch event.Topic {
	case debug.PocketTopic_CONSENSUS_MESSAGE_TOPIC:
		return node.GetBus().GetConsensusModule().HandleMessage(event.Data)
	case debug.PocketTopic_DEBUG_TOPIC:
		return node.handleDebugEvent(event.Data)
	case debug.PocketTopic_POCKET_NODE_TOPIC:
		log.Println("[NOOP] Received pocket node topic signal")
	default:
		log.Printf("[WARN] Unsupported PocketEvent topic: %s \n", event.Topic)
	}
	return nil
}

func (node *Node) handleDebugEvent(anyMessage *anypb.Any) error {
	var debugMessage debug.DebugMessage
	err := anypb.UnmarshalTo(anyMessage, &debugMessage, proto.UnmarshalOptions{})
	if err != nil {
		return err
	}
	switch debugMessage.Action {
	// Consensus Debug
	case debug.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
		fallthrough
	case debug.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE:
		fallthrough
	case debug.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW:
		fallthrough
	case debug.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE:
		return node.GetBus().GetConsensusModule().HandleDebugMessage(&debugMessage)
	// Persistence Debug
	case debug.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		fallthrough
	case debug.DebugMessageAction_DEBUG_PERSISTENCE_TREE_EXPORT:
		return node.GetBus().GetPersistenceModule().HandleDebugMessage(&debugMessage)
	default:
		log.Printf("Debug message: %s \n", debugMessage.Message)
	}

	return nil
}

func (node *Node) GetModuleName() string {
	return MainModuleName
}

func (node *Node) GetP2PAddress() cryptoPocket.Address {
	return node.p2pAddress
}
