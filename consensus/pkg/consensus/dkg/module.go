package dkg

import (
	"log"
	"pocket/consensus/pkg/config"

	"pocket/consensus/pkg/types"
	"pocket/shared/context"
	"pocket/shared/modules"

	"github.com/coinbase/kryptology/pkg/dkg/gennaro"
	v1 "github.com/coinbase/kryptology/pkg/sharing/v1"
)

type DKGModule interface {
	modules.PocketModule

	HandleMessage(*context.PocketContext, *DKGMessage)
}

type dkgModule struct {
	DKGModule
	pocketBusMod modules.PocketBusModule

	NodeId types.NodeId

	DKGParticipant      *gennaro.Participant
	ThresholdSigningKey *v1.ShamirShare

	// TODO: Move this over to the persistence module.
	DKGMessagePool map[DKGRound][]DKGMessage
}

func Create(
	cfg *config.Config,
) (m DKGModule, err error) {
	log.Println("Creating dkg module")

	consensusConfig := cfg.Consensus
	m = &dkgModule{
		NodeId: consensusConfig.NodeId,

		DKGParticipant:      nil,
		ThresholdSigningKey: nil,
		DKGMessagePool: make(map[DKGRound][]DKGMessage),
	}
	return m, nil
}

func (m *dkgModule) Start(ctx *context.PocketContext) error {
	log.Println("Starting dkg module")

	// TODO: Is this the right place for this?
	if m.DKGParticipant == nil {
		log.Printf("[DEBUG][%d] DKG Participant is nil, so adding a new one.\n", m.NodeId)
		m.addNewDKGParticipant()
	}
	return nil
}

func (m *dkgModule) Stop(ctx *context.PocketContext) error {
	log.Println("Stopping dkg module")
	return nil
}
