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

// TODO(#509): improve the documentation for this and other interfaces/functions
// ModuleOption is a function that configures a module when it is created.
// It uses a widely used pattern in Go called functional options.
// See https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
// for more information.
//
// It is used to provide optional parameters to the module constructor for all the cases
// where there is no configuration, which is often the case for sub-modules that are used
// and configured at runtime.
//
// It accepts an InitializableModule as a parameter, because in order to create a module with these options,
// at a minimum, the module must implement the InitializableModule interface.
//
// Example:
//
//	func WithFoo(foo string) ModuleOption {
//	  return func(m InitializableModule) {
//	    m.(*MyModule).foo = foo
//	  }
//	}
//
//	func NewMyModule(options ...ModuleOption) (Module, error) {
//	  m := &MyModule{}
//	  for _, option := range options {
//	    option(m)
//	  }
//	  return m, nil
//	}
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
