package service

import (
	"github.com/pokt-network/pocket/shared/modules"
)

const servicerModuleName = "servicer"

var _ modules.Module = &servicer{}
