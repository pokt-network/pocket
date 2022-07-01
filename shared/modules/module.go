package modules

import "github.com/pokt-network/pocket/shared/logging"

// TODO(olshansky): Show an example of `TypicalUsage`
type Module interface {
	Start() error
	Stop() error

	SetBus(Bus)
	GetBus() Bus

	GetLogger() logging.Logger
}
