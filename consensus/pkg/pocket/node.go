package pocket

import (

	// "crypto/ed25519"

	"log"

	"pocket/consensus/pkg/config"
	"pocket/consensus/pkg/consensus"
	"pocket/consensus/pkg/p2p"
	"pocket/consensus/pkg/persistance"
	"pocket/consensus/pkg/shared"
	"pocket/consensus/pkg/shared/context"
	"pocket/consensus/pkg/shared/modules"
	"pocket/consensus/pkg/types"

	"pocket/consensus/pkg/utility"
)

// TODO: SHould we create an interface for this as well?
type PocketNode struct {
	*modules.BasePocketModule

	// TODO: Should we export these fields or create getters?
	PersistanceMod modules.PersistanceModule
	NetworkMod     modules.NetworkModule
	UtilityMod     modules.UtilityModule
	ConsensusMod   modules.ConsensusModule

	Address string
}

func Create(ctx *context.PocketContext, config *config.Config) (n *PocketNode, err error) {
	// TODO: This loads
	state := shared.GetPocketState()
	state.LoadStateFromConfig(config)

	baseModule, err := modules.NewBaseModule(ctx, config)
	if err != nil {
		return nil, err
	}

	persistanceMod, err := persistance.Create(ctx, baseModule)
	if err != nil {
		return nil, err
	}

	networkMod, err := p2p.Create(ctx, baseModule)
	if err != nil {
		return nil, err
	}

	utilityMod, err := utility.Create(ctx, baseModule)
	if err != nil {
		return nil, err
	}

	// TODO: I don't like how the Create func signature is different depending on module dependency.
	// Need to think about how to leverage event bus versus direct calls.
	consensusMod, err := consensus.Create(ctx, baseModule)
	if err != nil {
		return nil, err
	}

	pocketBus, err := shared.CreatePocketBus(nil, persistanceMod, networkMod, utilityMod, consensusMod)
	if err != nil {
		return nil, err
	}

	baseModule.SetPocketBusMod(pocketBus)

	return &PocketNode{
		BasePocketModule: baseModule,

		PersistanceMod: persistanceMod,
		NetworkMod:     networkMod,
		UtilityMod:     utilityMod,
		ConsensusMod:   consensusMod,

		Address: types.AddressFromKey(config.PrivateKey.Public()),
	}, nil
}

func (node *PocketNode) Start(ctx *context.PocketContext) error {
	log.Println("Starting pocket node...")

	if err := node.PersistanceMod.Start(ctx); err != nil {
		return err
	}

	if err := node.NetworkMod.Start(ctx); err != nil {
		return err
	}

	if err := node.UtilityMod.Start(ctx); err != nil {
		return err
	}

	if err := node.ConsensusMod.Start(ctx); err != nil {
		return err
	}

	for {
		event := node.GetPocketBusMod().GetBusEvent()
		if err := node.handleEvent(event); err != nil {
			log.Println("Error handling event: ", err)
		}
	}
}
