package base_modules

import "github.com/pokt-network/pocket/shared/modules"

var _ modules.InterruptableModule = &InterruptableModule{}

// InterruptableModule is a noop implementation of the InterruptableModule interface.
//
// It is useful for modules that do not need any particular logic to be executed when started or stopped.
// In these situations, just embed this struct into the module struct.
type InterruptableModule struct{}

func (*InterruptableModule) Start() error {
	return nil
}

func (*InterruptableModule) Stop() error {
	return nil
}
