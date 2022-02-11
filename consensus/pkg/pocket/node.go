package pocket

import (

	// "crypto/ed25519"

	"log"

	"pocket/consensus/pkg/config"
	"pocket/consensus/pkg/consensus"
	"pocket/p2p"
	"pocket/persistence"

	"pocket/consensus/pkg/types"
	"pocket/shared"
	"pocket/shared/context"
	"pocket/shared/modules"

	"pocket/utility/utility"
)

// TODO: SHould we create an interface for this as well?
type PocketNode struct {
	modules.PocketModule

	pocketBusMod modules.PocketBusModule

	Address string
}

func Create(ctx *context.PocketContext, config *config.Config) (n *PocketNode, err error) {
	// TODO: This loads
	state := shared.GetPocketState()
	state.LoadStateFromConfig(config)

	// baseModule, err := modules.NewBaseModule(ctx, config)
	// if err != nil {
	// 	return nil, err
	// }

	persistenceMod, err := persistence.Create(config)
	if err != nil {
		return nil, err
	}

	networkMod, err := p2p.Create(config)
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

	bus, err := shared.CreatePocketBus(nil, persistenceMod, networkMod, utilityMod, consensusMod)
	if err != nil {
		return nil, err
	}

	return &PocketNode{
		pocketBusMod: bus,
		Address:      types.AddressFromKey(config.PrivateKey.Public()),
	}, nil
}

func (node *PocketNode) Start(ctx *context.PocketContext) error {
	log.Println("Starting pocket node...")

	// NOTE: Order of module initializaiton matters.

	if err := node.GetPocketBusMod().GetPersistenceModule().Start(ctx); err != nil {
		return err
	}

	if err := node.GetPocketBusMod().GetNetworkModule().Start(ctx); err != nil {
		return err
	}

	//if err := node.GetPocketBusMod().GetUtilityModule().Start(ctx); err != nil {
	//	return err
	//}

	if err := node.GetPocketBusMod().GetConsensusModule().Start(ctx); err != nil {
		return err
	}

	for {
		event := node.GetPocketBusMod().GetBusEvent()
		if err := node.handleEvent(event); err != nil {
			log.Println("Error handling event: ", err)
		}
	}
}

func (m *PocketNode) SetPocketBusMod(pocketBus modules.PocketBusModule) {
	m.pocketBusMod = pocketBus
}

func (m *PocketNode) GetPocketBusMod() modules.PocketBusModule {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}
