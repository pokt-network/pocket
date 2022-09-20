package shared

import (
	"log"

	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared/debug"
	"github.com/pokt-network/pocket/telemetry"

	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/persistence"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.Module = &Node{}

const (
	MainModuleName = "main"
)

type Node struct {
	bus     modules.Bus
	Address cryptoPocket.Address
}

func Create(configPath, genesisPath string) (n *Node, err error) {
	runtime := runtime.NewBuilder(configPath, genesisPath)
	cfg := runtime.GetConfig()
	genesis := runtime.GetGenesis()

	persistenceMod, err := persistence.Create(cfg.Persistence, genesis.PersistenceGenesisState)
	if err != nil {
		return nil, err
	}

	p2pMod, err := p2p.Create(cfg.P2P, false)
	if err != nil {
		return nil, err
	}

	utilityMod, err := utility.Create(cfg.Utility)
	if err != nil {
		return nil, err
	}

	consensusMod, err := consensus.Create(runtime, false)
	if err != nil {
		return nil, err
	}

	telemetryMod, err := telemetry.Create(cfg.Telemetry)
	if err != nil {
		return nil, err
	}

	bus, err := CreateBus(cfg, genesis, persistenceMod, p2pMod, utilityMod, consensusMod.(modules.ConsensusModule), telemetryMod)
	if err != nil {
		return nil, err
	}
	addr, err := p2pMod.GetAddress()
	if err != nil {
		return nil, err
	}
	return &Node{
		bus:     bus,
		Address: addr,
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
	case debug.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
		fallthrough
	case debug.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE:
		fallthrough
	case debug.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW:
		fallthrough
	case debug.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE:
		return node.GetBus().GetConsensusModule().HandleDebugMessage(&debugMessage)
	case debug.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		return node.GetBus().GetPersistenceModule().HandleDebugMessage(&debugMessage)
	default:
		log.Printf("Debug message: %s \n", debugMessage.Message)
	}

	return nil
}

func (node *Node) GetModuleName() string {
	return MainModuleName
}
