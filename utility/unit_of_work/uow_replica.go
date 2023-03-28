package unit_of_work

import (
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ modules.UtilityUnitOfWork        = &replicaUtilityUnitOfWork{}
	_ modules.ReplicaUtilityUnitOfWork = &replicaUtilityUnitOfWork{}
)

type replicaUtilityUnitOfWork struct {
	baseUtilityUnitOfWork
}

func NewReplicaUOW(height int64, readContext modules.PersistenceReadContext, rwPersistenceContext modules.PersistenceRWContext) *replicaUtilityUnitOfWork {
	return &replicaUtilityUnitOfWork{
		baseUtilityUnitOfWork: baseUtilityUnitOfWork{
			height:                 height,
			persistenceReadContext: readContext,
			persistenceRWContext:   rwPersistenceContext,
			logger:                 logger.Global.CreateLoggerForModule(replicaUtilityUOWModuleName),
		},
	}
}
