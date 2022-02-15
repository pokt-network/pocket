package dkg

import (
	"log"
	"pocket/consensus/types"
	"pocket/shared/config"
	"pocket/shared/modules"

	"github.com/coinbase/kryptology/pkg/dkg/gennaro"
	v1 "github.com/coinbase/kryptology/pkg/sharing/v1"
)

type DKGModule interface {
	modules.Module

	HandleMessage(*DKGMessage)
}

type dkgModule struct {
	DKGModule
	pocketBusMod modules.BusModule

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
		NodeId: types.NodeId(consensusConfig.NodeId),

		DKGParticipant:      nil,
		ThresholdSigningKey: nil,
		DKGMessagePool:      make(map[DKGRound][]DKGMessage),
	}
	return m, nil
}

func (m *dkgModule) Start() error {
	log.Println("Starting dkg module")

	// TODO: Is this the right place for this?
	if m.DKGParticipant == nil {
		log.Printf("[DEBUG][%d] DKG Participant is nil, so adding a new one.\n", m.NodeId)
		m.addNewDKGParticipant()
	}
	return nil
}

func (m *dkgModule) Stop() error {
	log.Println("Stopping dkg module")
	return nil
}
