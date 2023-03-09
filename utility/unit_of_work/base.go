package unit_of_work

import (
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

const (
	leaderUtilityUOWModuleName  = "leader_utility_unit_of_work"
	replicaUtilityUOWModuleName = "replica_utility_unit_of_work"
)

var _ modules.UtilityUnitOfWork = &baseUtilityUnitOfWork{}

type baseUtilityUnitOfWork struct {
	base_modules.IntegratableModule

	logger *modules.Logger

	persistenceReadContext modules.PersistenceReadContext
	persistenceRWContext   modules.PersistenceRWContext

	// TECHDEBT: Consolidate all these types with the shared Protobuf struct and create a `proposalBlock`
	proposalStateHash    string
	proposalProposerAddr []byte
	proposalBlockTxs     [][]byte
}

func (uow *baseUtilityUnitOfWork) SetProposalBlock(blockHash string, proposerAddr []byte, txs [][]byte) error {
	uow.proposalStateHash = blockHash
	uow.proposalProposerAddr = proposerAddr
	uow.proposalBlockTxs = txs
	return nil
}

func (uow *baseUtilityUnitOfWork) ApplyBlock(beforeApplyBlock, afterApplyBlock func(modules.UtilityUnitOfWork) error) (stateHash string, txs [][]byte, err error) {
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

func (uow *baseUtilityUnitOfWork) Commit(quorumCert []byte) error {
	// TODO: @deblasis - change tracking here

	uow.logger.Debug().Msg("committing the rwPersistenceContext...")
	if err := uow.persistenceRWContext.Commit(uow.proposalProposerAddr, quorumCert); err != nil {
		return err
	}
	return uow.Release()
}

func (uow *baseUtilityUnitOfWork) Release() error {
	// TODO: @deblasis - change tracking reset here

	if uow.persistenceRWContext == nil {
		return nil
	}
	if err := uow.persistenceRWContext.Release(); err != nil {
		return err
	}
	if err := uow.persistenceReadContext.Close(); err != nil {
		return err
	}
	return nil
}
