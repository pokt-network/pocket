package modules

import (
	"github.com/pokt-network/pocket/shared/crypto"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

type Module interface {
	InitializableModule
	IntegratableModule
	InterruptableModule
}

type IntegratableModule interface {
	SetBus(Bus)
	GetBus() Bus
}

type InterruptableModule interface {
	Start() error
	Stop() error
}

type InitializableModule interface {
	GetModuleName() string
	Create(runtime RuntimeMgr) (Module, error)
}

type ConfigurableModule interface {
	ValidateConfig(Config) error
}

type GenesisDependentModule interface {
	ValidateGenesis(GenesisState) error
}

type KeyholderModule interface {
	GetPrivateKey() (crypto.PrivateKey, error)
}

type P2PAddressableModule interface {
	GetP2PAddress() cryptoPocket.Address
}

// TODO(@okdas): this should also initialize metrics-related stuff, just the logger for now
// InitMetrics(pathToConfigJSON string)
// InitTracing(pathToConfigJSON string)
type ObservableModule interface {
	InitLogger(pathToConfigJSON string)
	GetLogger() Logger
}
