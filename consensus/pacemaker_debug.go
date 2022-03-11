package consensus

import (
	"time"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

// This Pacemaker interface is only used for development & debugging purposes.
type PacemakerDebug interface {
	SetManualMode(bool)
	IsManualMode() bool
	ForceNextView()
}

type paceMakerDebug struct {
	manualMode                bool
	debugTimeBetweenStepsMsec uint64

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

// This is a hack only used to slow down the progress of the blockchain during development.
func (p *paceMaker) debugSleep() {
	time.Sleep(time.Duration(int64(time.Millisecond) * int64(p.debugTimeBetweenStepsMsec)))
}
