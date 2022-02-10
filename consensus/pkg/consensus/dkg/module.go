package dkg

import (
	"log"

	"pocket/consensus/pkg/shared/context"
	"pocket/consensus/pkg/shared/modules"
	"pocket/consensus/pkg/types"

	"github.com/coinbase/kryptology/pkg/dkg/gennaro"
	v1 "github.com/coinbase/kryptology/pkg/sharing/v1"
)

type DKGModule interface {
	modules.PocketModule

	HandleMessage(*context.PocketContext, *DKGMessage)
}

type dkgModule struct {
	*modules.BasePocketModule

	NodeId types.NodeId

	DKGParticipant      *gennaro.Participant
	ThresholdSigningKey *v1.ShamirShare

	// TODO: Move this over to the persistance module.
	DKGMessagePool map[DKGRound][]DKGMessage
}

func Create(
	ctx *context.PocketContext,
	base *modules.BasePocketModule,
) (m DKGModule, err error) {
	log.Println("Creating dkg module")
	m = &dkgModule{
		BasePocketModule: base,

		NodeId: base.GetConfig().Consensus.NodeId,

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
