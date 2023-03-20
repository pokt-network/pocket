package utility

import (
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility/unit_of_work"
)

func (u *utilityModule) NewUnitOfWork(height int64) (modules.UtilityUnitOfWork, error) {
	readContext, err := u.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return nil, err
	}
	rwContext, err := u.GetBus().GetPersistenceModule().NewRWContext(height)
	if err != nil {
		return nil, err
	}

	var utilityUOW modules.UtilityUnitOfWork
	if u.GetBus().GetConsensusModule().IsLeader() {
		utilityUOW = unit_of_work.NewForLeader(height, readContext, rwContext)
	} else {
		utilityUOW = unit_of_work.NewForReplica(height, readContext, rwContext)
	}

	utilityUOW.SetBus(u.GetBus())
	return utilityUOW, nil
}
