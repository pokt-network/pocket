package unit_of_work

import (
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ modules.UtilityUnitOfWork       = &leaderUtilityUnitOfWork{}
	_ modules.LeaderUtilityUnitOfWork = &leaderUtilityUnitOfWork{}
)

type leaderUtilityUnitOfWork struct {
	baseUtilityUnitOfWork
}

func NewForLeader(height int64, readContext modules.PersistenceReadContext, rwPersistenceContext modules.PersistenceRWContext) *leaderUtilityUnitOfWork {
	return &leaderUtilityUnitOfWork{
		baseUtilityUnitOfWork: baseUtilityUnitOfWork{
			persistenceReadContext: readContext,
			persistenceRWContext:   rwPersistenceContext,
			logger:                 logger.Global.CreateLoggerForModule(leaderUtilityUOWModuleName),
		},
	}
}

func (uow *leaderUtilityUnitOfWork) CreateProposalBlock(proposer []byte, maxTxBytes uint64, beforeApplyBlock, afterApplyBlock func(modules.UtilityUnitOfWork) error) (stateHash string, txs [][]byte, err error) {
	if beforeApplyBlock != nil {
		uow.logger.Debug().Msg("running beforeApplyBlock...")
		if err := beforeApplyBlock(uow); err != nil {
			return "", nil, err
		}
	}

	if afterApplyBlock != nil {
		uow.logger.Debug().Msg("running afterApplyBlock...")
		if err := afterApplyBlock(uow); err != nil {
			return "", nil, err
		}
	}
	panic("unimplemented")
}
