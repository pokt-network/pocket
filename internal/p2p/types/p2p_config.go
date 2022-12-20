package types

import "github.com/pokt-network/pocket/internal/shared/modules"

var _ modules.P2PConfig = &P2PConfig{}

func (x *P2PConfig) IsEmptyConnType() bool {
	if x.GetIsEmptyConnectionType() {
		return true
	}
	return false
}
