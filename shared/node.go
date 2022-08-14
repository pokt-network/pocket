package shared

import (
	"github.com/pokt-network/pocket/shared/types/genesis"
	"log"

	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/persistence"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/shared/types"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/telemetry"
)

var _ modules.Module = &Node{}

type Node struct {
	bus     modules.Bus
	Address cryptoPocket.Address
}

func Create(cfg *genesis.Config, genesis *genesis.GenesisState) (n *Node, err error) {
	persistenceMod, err := persistence.Create(cfg, genesis)
	if err != nil {
		return nil, err
	}

	p2pMod, err := p2p.Create(cfg, genesis)
	if err != nil {
		return nil, err
	}

	utilityMod, err := utility.Create(cfg, genesis)
	if err != nil {
		return nil, err
	}

	consensusMod, err := consensus.Create(cfg, genesis)
	if err != nil {
		return nil, err
	}

	telemetryMod, err := telemetry.Create(cfg, genesis) // TODO (team; discuss) is telemetry a proper module or not?
	if err != nil {
		return nil, err
	}

	bus, err := CreateBus(
		persistenceMod,
		p2pMod,
		utilityMod,
		consensusMod,
		telemetryMod,
		cfg,
	)

	if err != nil {
		return nil, err
	}

	pk, err := cryptoPocket.NewPrivateKey(cfg.Base.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &Node{
		bus:     bus,
		Address: pk.Address(),
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
	signalNodeStartedEvent := &types.PocketEvent{Topic: types.PocketTopic_POCKET_NODE_TOPIC, Data: nil}
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

func (node *Node) handleEvent(event *types.PocketEvent) error {
	switch event.Topic {
	case types.PocketTopic_CONSENSUS_MESSAGE_TOPIC:
		return node.GetBus().GetConsensusModule().HandleMessage(event.Data)
	case types.PocketTopic_DEBUG_TOPIC:
		return node.handleDebugEvent(event.Data)
	case types.PocketTopic_POCKET_NODE_TOPIC:
		log.Println("[NOOP] Received pocket node topic signal")
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
	case types.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		return node.GetBus().GetPersistenceModule().HandleDebugMessage(&debugMessage)
	default:
		log.Printf("Debug message: %s \n", debugMessage.Message)
	}

	return nil
}
