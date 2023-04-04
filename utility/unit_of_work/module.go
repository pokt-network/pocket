package unit_of_work

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	utilTypes "github.com/pokt-network/pocket/utility/types"
)

const (
	leaderUtilityUOWModuleName  = "leader_utility_UOW"
	replicaUtilityUOWModuleName = "replica_utility_UOW"
)

var _ modules.UtilityUnitOfWork = &baseUtilityUnitOfWork{}

type baseUtilityUnitOfWork struct {
	base_modules.IntegratableModule

	logger *modules.Logger

	height int64

	// TECHDEBT(#564): the way we access the contexts and apply changes to them is still a work in progress.
	// The path forward will become clearer during the implementation of change tracking in #564.
	// For now, it seems sensible to have separate contexts for read and write operations.
	// The idea is that:
	// - the write context will track ephemeral changes and also provide a way to persist them.
	// - the read context will only read but also see what has been changed in the ephemeral state.
	// from the consumers of the unit of work point of view, this is just an implementation detail.
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

// CLEANUP: code re-use ApplyBlock() for CreateAndApplyBlock()
func (u *baseUtilityUnitOfWork) ApplyBlock() (string, [][]byte, error) {
	if !u.isProposalBlockSet() {
		return "", nil, utilTypes.ErrProposalBlockNotSet()
	}
	// begin block lifecycle phase
	if err := u.beginBlock(); err != nil {
		return "", nil, err
	}

	mempool := u.GetBus().GetUtilityModule().GetMempool()

	// deliver txs lifecycle phase
	for index, txProtoBytes := range u.proposalBlockTxs {
		tx, err := coreTypes.TxFromBytes(txProtoBytes)
		if err != nil {
			return "", nil, err
		}
		if err := tx.ValidateBasic(); err != nil {
			return "", nil, err
		}
		// TODO(#346): Currently, the pattern is allowing nil err with an error transaction...
		//             Should we terminate applyBlock immediately if there's an invalid transaction?
		//             Or wait until the entire lifecycle is over to evaluate an 'invalid' block

		// Validate and apply the transaction to the Postgres database
		txResult, err := u.hydrateTxResult(tx, index)
		if err != nil {
			return "", nil, err
		}

		txHash, err := tx.Hash()
		if err != nil {
			return "", nil, err
		}

		// TODO: Need to properly add transactions back on rollbacks
		if mempool.Contains(txHash) {
			if err := mempool.RemoveTx(txProtoBytes); err != nil {
				return "", nil, err
			}
			u.logger.Info().Str("tx_hash", txHash).Msg("Applying tx that WAS in the local mempool")
		} else {
			u.logger.Info().Str("tx_hash", txHash).Msg("Applying tx that WAS NOT in the local mempool")
		}

		if err := u.persistenceRWContext.IndexTransaction(txResult); err != nil {
			u.logger.Fatal().Err(err).Msgf("TODO(#327): We can apply the transaction but not index it. Crash the process for now: %v\n", err)
		}
	}

	// end block lifecycle phase
	if err := u.endBlock(u.proposalProposerAddr); err != nil {
		return "", nil, err
	}
	// TODO(@deblasis): this should be from a ReadContext (the ephemeral/staging one)
	// return the app hash (consensus module will get the validator set directly)
	stateHash, err := u.persistenceRWContext.ComputeStateHash()
	if err != nil {
		u.logger.Fatal().Err(err).Msg("Updating the app hash failed. TODO: Look into roll-backing the entire commit...")
		return "", nil, utilTypes.ErrAppHash(err)
	}
	u.logger.Info().Str("state_hash", stateHash).Msgf("ApplyBlock succeeded!")

	// return the app hash; consensus module will get the validator set directly
	return stateHash, nil, nil
}

// TODO(@deblasis): change tracking here
func (uow *baseUtilityUnitOfWork) Commit(quorumCert []byte) error {
	uow.logger.Debug().Msg("committing the rwPersistenceContext...")
	if err := uow.persistenceRWContext.Commit(uow.proposalProposerAddr, quorumCert); err != nil {
		return err
	}
	uow.persistenceRWContext = nil
	return nil
}

// TODO(@deblasis): change tracking reset here
func (uow *baseUtilityUnitOfWork) Release() error {
	rwCtx := uow.persistenceRWContext
	if rwCtx != nil {
		uow.persistenceRWContext = nil
		rwCtx.Release()
	}

	readCtx := uow.persistenceReadContext
	if readCtx != nil {
		uow.persistenceReadContext = nil
		readCtx.Release()
	}

	return nil
}

func (uow *baseUtilityUnitOfWork) isProposalBlockSet() bool {
	return uow.proposalStateHash != "" && uow.proposalProposerAddr != nil && uow.proposalBlockTxs != nil
}
