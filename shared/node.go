package shared

import (
	"log"

	"github.com/pokt-network/pocket/p2p/pre2p"
	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility"

	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/types"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.Module = &Node{}

type Node struct {
	bus modules.Bus

	Address cryptoPocket.Address
}

func Create(cfg *config.Config) (n *Node, err error) {
	// TODO(design): initialize the state singleton until we have a proper solution for this.
	_ = types.GetTestState(cfg)

	persistenceMod, err := persistence.Create(cfg)
	if err != nil {
		return nil, err
	}

	// TODO(derrandz): Replace with real P2P module
	// p2pMod, err := p2p.Create(cfg)
	pre2pMod, err := pre2p.Create(cfg)
	if err != nil {
		return nil, err
	}

	// TODO(andrew): Replace with real Utility module
	// utilityMod, err := utility.Create(cfg)
	mockedUtilityMod, err := utility.CreateMockedModule(cfg)
	if err != nil {
		return nil, err
	}

	consensusMod, err := consensus.Create(cfg)
	if err != nil {
		return nil, err
	}

	bus, err := CreateBus(persistenceMod, pre2pMod, mockedUtilityMod, consensusMod)
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
