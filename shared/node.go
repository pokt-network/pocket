package shared

import (
	"fmt"
	"log"
	"pocket/consensus"
	"pocket/persistence"
	"pocket/pre2p"
	"pocket/shared/config"
	"pocket/shared/crypto"
	"pocket/shared/types"
	"pocket/utility"
	"time"

	"pocket/shared/modules"
)

var _ modules.Module = &Node{}

type Node struct {
	bus modules.Bus

	Address string
}

func Create(config *config.Config) (n *Node, err error) {
	pk, err := crypto.NewPrivateKey(config.PrivateKey)
	if err != nil {
		return nil, err
	}

	persistenceMod, err := persistence.Create(config)
	if err != nil {
		return nil, err
	}

	// TODO(derrands): Replace with real P2P module
	pre2pMod, err := pre2p.Create(config)
	if err != nil {
		return nil, err
	}

	utilityMod, err := utility.Create(config)
	if err != nil {
		return nil, err
	}

	consensusMod, err := consensus.Create(config)
	if err != nil {
		return nil, err
	}

	bus, err := CreateBus(nil, persistenceMod, pre2pMod, utilityMod, consensusMod)
	if err != nil {
		return nil, err
	}

	return &Node{
		bus:     bus,
		Address: pk.Address().String(),
	}, nil
}

func (node *Node) Start() error {
	log.Println("Starting pocket node...")

	// NOTE: Order of module initialization matters.

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

	// TODO(olshansky): This is a temporary function in order to avoid
	// the `all goroutines are asleep - deadlock` error. This happens because
	// there is no existing functionality in the source code, so the static
	// compiler understands that no future events will take place.
	go func() {
		for {
			fmt.Println("Sending placeholder event...")
			node.GetBus().PublishEventToBus(&types.Event{PocketTopic: "Placeholder"})
			time.Sleep(time.Second * 5)
		}
	}()

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
