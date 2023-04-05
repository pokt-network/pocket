package savepoints

import (
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	moduleName = "savepoint_factory"

	// appsKey       = "apps"
	// validatorsKey = "validators"
	// fishermenKey  = "fishermen"
	// servicersKey  = "servicers"
	// accountsKey   = "accounts"
	// poolsKey      = "pools"
	// paramsKey     = "params"
	// flagsKey      = "flags"
)

var _ modules.SavepointFactory = &savepointFactory{}

type savepointFactory struct {
	readContext modules.PersistenceReadContext
	logger      *modules.Logger
}

func NewSavepointFactory(readContext modules.PersistenceReadContext) modules.SavepointFactory {
	return &savepointFactory{
		readContext: readContext,
		logger:      logger.Global.CreateLoggerForModule(moduleName),
	}
}

func (sm *savepointFactory) CreateSavepoint(height int64) (modules.PersistenceReadContext, error) {
	sm.logger.Debug().Bool("TODO", true).Msg("unimplemented")
	return &savepoint{
		height: height,
	}, nil
}
