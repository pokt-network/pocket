package shared

import (
	"log"

	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/persistence/pre_persistence"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/telemetry"
)

var _ modules.Module = &Node{}

type Node struct {
	bus     modules.Bus
	Address cryptoPocket.Address
}

func Create(cfg *config.Config) (n *Node, err error) {
	// TODO(design): initialize the state singleton until we have a proper solution for this.
	_ = typesGenesis.GetNodeState(cfg)

	// TODO(drewsky): The module is initialized to run background processes during development
	// to make sure it's part of the node's lifecycle, but is not referenced YET byt the app specific
	// bus.
	if _, err := persistence.Create(cfg); err != nil {
		return nil, err
	}

	// TODO(drewsky): deprecate pre-persistence
	prePersistenceMod, err := pre_persistence.Create(cfg)
	if err != nil {
		return nil, err
	}
	// TODO(derrandz): Replace with real P2P module
	// p2pMod, err := p2p.Create(cfg)
	// pre2pMod, err := pre2p.Create(cfg)
	p2pMod, err := p2p.Create(cfg)
	if err != nil {
		return nil, err
	}

	// TODO(andrew): Replace with real Utility module
	utilityMod, err := utility.Create(cfg)
	// mockedUtilityMod, err := utility.CreateMockedModule(cfg)
	if err != nil {
		return nil, err
	}

	consensusMod, err := consensus.Create(cfg)
	if err != nil {
		return nil, err
	}

	telemetryMod, err := telemetry.Create(cfg)
	if err != nil {
		return nil, err
	}

	bus, err := CreateBus(prePersistenceMod, p2pMod, utilityMod, consensusMod, telemetryMod)
	if err != nil {
		return nil, err
	}

	return &Node{
		bus:     bus,
		Address: cfg.PrivateKey.Address(),
	}, nil
}

func (node *Node) Start() error {
	log.Println("Starting pocket node...")

	// NOTE: Order of module startup here matters.

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

	// TODO(olshansky): discuss if we need a special type/event for this.
	signalNodeStartedEvent := &types.PocketEvent{Topic: types.PocketTopic_POCKET_NODE_TOPIC, Data: nil}
	node.GetBus().PublishEventToBus(signalNodeStartedEvent)

	// While loop lasting throughout the entire lifecycle of the node.
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

func (node *Node) handleEvent(event *types.PocketEvent) error {
	switch event.Topic {
	case types.PocketTopic_CONSENSUS_MESSAGE_TOPIC:
		return node.GetBus().GetConsensusModule().HandleMessage(event.Data)
	case types.PocketTopic_DEBUG_TOPIC:
		return node.handleDebugEvent(event.Data)
	case types.PocketTopic_POCKET_NODE_TOPIC:
		log.Println("NOOP - Received pocket node topic signal")
	default:
		log.Printf("[WARN] Unsupported PocketEvent topic: %s \n", event.Topic)
	}
	return nil
}

func (node *Node) handleDebugEvent(anyMessage *anypb.Any) error {
	var debugMessage types.DebugMessage
	err := anypb.UnmarshalTo(anyMessage, &debugMessage, proto.UnmarshalOptions{})
	if err != nil {
		return err
	}
	switch debugMessage.Action {
	case types.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
		fallthrough
	case types.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE:
		fallthrough
	case types.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW:
		fallthrough
	case types.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE:
		return node.GetBus().GetConsensusModule().HandleDebugMessage(&debugMessage)
	default:
		log.Printf("Debug message: %s \n", debugMessage.Message)
	}

	return nil
}
