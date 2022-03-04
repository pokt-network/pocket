package consensus

import (
	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

// TODO: PaceMaker has some functions that are meant only part of the interface
// for development and debugging purposes. Need to think about how to decouple
// it (if needed) to avoid code complexity in the core business logic.

type PaceMakerDebug interface {
	SetManualMode(bool)
	IsManualMode() bool
	ForceNextView()
}

type paceMakerDebug struct {
	// Debug variables.
	manualMode        bool
	quorumCertificate *types_consensus.QuorumCertificate
}

func (p *paceMaker) IsManualMode() bool {
	return p.manualMode
}

func (p *paceMaker) SetManualMode(manualMode bool) {
	p.manualMode = manualMode
}

func (p *paceMaker) ForceNextView() {
	lastQC := p.quorumCertificate
	p.startNextView(lastQC, true)
}
