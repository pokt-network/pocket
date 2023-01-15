package pacemaker

import typesCons "github.com/pokt-network/pocket/consensus/types"

var _ PacemakerDebug = &pacemaker{}

// This Pacemaker interface is only used for development & debugging purposes.
type PacemakerDebug interface {
	SetManualMode(bool)
	IsManualMode() bool
	ForceNextView()
}

type pacemakerdebug struct {
	manualMode                bool
	debugTimeBetweenStepsMsec uint64
	quorumCertificate         *typesCons.QuorumCertificate
}

func (p *pacemaker) IsManualMode() bool {
	return p.manualMode
}

func (p *pacemaker) SetManualMode(manualMode bool) {
	p.manualMode = manualMode
}

func (p *pacemaker) ForceNextView() {
	p.startNextView(p.quorumCertificate, true)
}
