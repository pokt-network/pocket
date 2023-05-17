package modules

import (
	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/genesis"
)

//go:generate mockgen -destination=./mocks/runtime_module_mock.go github.com/pokt-network/pocket/shared/modules RuntimeMgr

type RuntimeMgr interface {
	GetConfig() *configs.Config
	GetGenesis() *genesis.GenesisState
	GetClock() clock.Clock
}
