package modules

import (
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

type ModuleOption func(InitializableModule)
type InitializableModule interface {
	GetModuleName() string
	Create(bus Bus, options ...ModuleOption) (Module, error)
}

type KeyholderModule interface {
	GetPrivateKey() (cryptoPocket.PrivateKey, error)
}

type P2PAddressableModule interface {
	GetP2PAddress() cryptoPocket.Address
}

type ObservableModule interface {
	GetLogger() Logger
}
