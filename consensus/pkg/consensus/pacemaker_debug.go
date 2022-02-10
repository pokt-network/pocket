package consensus

// TODO: PaceMaker has some functionsthat are meant only part of the interface
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
	quorumCertificate *QuorumCertificate
}

func (p *paceMaker) IsManualMode() bool {
	return p.manualMode
}

func (p *paceMaker) SetManualMode(manualMode bool) {
	p.manualMode = manualMode
}

func (p *paceMaker) ForceNextView() {
	lastQC := p.quorumCertificate
	p.InterruptRound()
	p.startNextView(lastQC, true)
}
