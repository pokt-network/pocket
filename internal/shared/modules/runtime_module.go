package modules

import "github.com/benbjohnson/clock"

//go:generate mockgen -source=$GOFILE -destination=./mocks/runtime_module_mock.go -aux_files=github.com/pokt-network/pocket/internal/shared/modules=module.go

type RuntimeMgr interface {
	GetConfig() Config
	GetGenesis() GenesisState
	GetClock() clock.Clock
}
