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
			height:                 height,
			persistenceReadContext: readContext,
			persistenceRWContext:   rwPersistenceContext,
			logger:                 logger.Global.CreateLoggerForModule(leaderUtilityUOWModuleName),
		},
	}
}

func (uow *leaderUtilityUnitOfWork) CreateProposalBlock(proposer []byte, maxTxBytes uint64) (stateHash string, txs [][]byte, err error) {
	// if burnFn != nil {
	// 	uow.logger.Debug().Msg("running burnFn...")
	// 	if err := burnFn(uow); err != nil {
	// 		return "", nil, err
	// 	}
	// }

	// if rewardFn != nil {
	// 	uow.logger.Debug().Msg("running rewardFn...")
	// 	if err := rewardFn(uow); err != nil {
	// 		return "", nil, err
	// 	}
	// }
	panic("unimplemented")
}
