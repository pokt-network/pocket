package pacemaker

import typesCons "github.com/pokt-network/pocket/consensus/types"

var (
	_ PacemakerDebug = &pacemaker{}
)

// This Pacemaker interface is only used for development & debugging purposes.
type PacemakerDebug interface {
	SetManualMode(bool)
	IsManualMode() bool
	ForceNextView()
}

type pacemakerDebug struct {
	manualMode                bool
	debugTimeBetweenStepsMsec uint64
	quorumCertificate         *typesCons.QuorumCertificate
}

func (m *pacemaker) IsManualMode() bool {
	return m.debug.manualMode
}

func (m *pacemaker) SetManualMode(manualMode bool) {
	m.debug.manualMode = manualMode
}

func (m *pacemaker) ForceNextView() {
	m.startNextView(m.debug.quorumCertificate, true)
}
