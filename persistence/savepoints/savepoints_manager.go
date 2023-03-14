package savepoints

import "github.com/pokt-network/pocket/shared/modules"

var _ modules.SavepointManager = &savepointManager{}

type savepointManager struct {
	readContext modules.PersistenceReadContext
}

func NewSavepointManager(readContext modules.PersistenceReadContext) modules.SavepointManager {
	return &savepointManager{
		readContext: readContext,
	}
}

// CreateSavepoint implements SavepointManager
func (*savepointManager) CreateSavepoint(height int64) modules.PersistenceReadContext {
	return &Savepoint{}
}
