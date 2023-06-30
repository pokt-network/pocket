package modules

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

type Module interface {
	InjectableModule
	IntegrableModule
	InterruptableModule
	ModuleFactoryWithOptions
}

type Submodule interface {
	InjectableModule
	IntegrableModule
}

// IntegrableModule is a module that integrates with the bus.
// Essentially it's a module that is capable of communicating with the `bus` (see `shared/modules/bus_module.go`) for additional details.
type IntegrableModule interface {
	// SetBus sets the bus for the module.
	//
	// Generally it is called by the `bus` itself whenever a module is registered via `bus.RegisterModule(module modules.Module)`
	SetBus(Bus)

	// GetBus returns the bus for the module.
	GetBus() Bus
}

// InjectableModule is a module that has some basic lifecycle logic. Specifically, it can be started and stopped.
type InterruptableModule interface {
	// Start starts the module and executes any logic that is required at the beginning of the module's lifecycle.
	Start() error

	// Stop stops the module and executes any logic that is required when the module's lifecycle is over.
	Stop() error
}

// ModuleOption is a function that configures a module when it is created.
// It uses a widely used pattern in Go called functional options.
// See https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
// for more information.
//
// It is used to provide optional parameters to the module constructor for all the cases
// where there is no configuration, which is often the case for sub-modules that are used
// and configured at runtime.
//
// It accepts an InjectableModule as a parameter, because in order to create a module with these options,
// at a minimum, the module must implement the InjectableModule interface.
//
// Example:
//
//	func WithFoo(foo string) ModuleOption {
//	  return func(m InjectableModule) {
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
type ModuleOption func(InjectableModule)

// InjectableModule is a module that can be created via the standardized `Create` method and that has a name
// that can be used to identify it (see `shared\modules\modules_registry_module.go`) for additional details.
type InjectableModule interface {
	// GetModuleName returns the name of the module.
	GetModuleName() string
}

// KeyholderModule is a module that can provide a private key.
type KeyholderModule interface {
	// GetPrivateKey returns the private key held by the module.
	GetPrivateKey() (cryptoPocket.PrivateKey, error)
}

// ObservableModule is a module that can provide observability via a Logger.
type ObservableModule interface {
	// GetLogger returns the logger for the module.
	GetLogger() Logger
}
