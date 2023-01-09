package debug

import "github.com/pokt-network/pocket/p2p/providers/current_height_provider"

var _ current_height_provider.CurrentHeightProvider = &debugCurrentHeightProvider{}

type debugCurrentHeightProvider struct {
	currentHeight uint64
}

func (dchp *debugCurrentHeightProvider) CurrentHeight() uint64 {
	return dchp.currentHeight
}

func NewDebugCurrentHeightProvider(height uint64) *debugCurrentHeightProvider {
	dchp := &debugCurrentHeightProvider{
		currentHeight: height,
	}

	return dchp
}
