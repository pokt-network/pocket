package modules

import (
	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/runtime/configs"
)

//go:generate mockgen -source=$GOFILE -destination=./mocks/runtime_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

type RuntimeMgr interface {
	GetConfig() *configs.Config
	GetGenesis() GenesisState
	GetClock() clock.Clock
}
