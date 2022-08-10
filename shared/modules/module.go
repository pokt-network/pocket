package modules

import "github.com/pokt-network/pocket/shared/logging"

// TODO(olshansky): Show an example of `TypicalUsage`
type Module interface {
	IntegratableModule
	InterruptableModule

	Loggable
}

type IntegratableModule interface {
	SetBus(Bus)
	GetBus() Bus
}

type InterruptableModule interface {
	Start() error
	Stop() error
}

type Loggable interface {
	Logger() logging.Logger
}
