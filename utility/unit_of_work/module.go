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

// CreateAndApplyProposalBlock implements the exposed functionality of the shared UtilityContext interface.
func (u *baseUtilityUnitOfWork) CreateAndApplyProposalBlock(proposer []byte, maxTransactionBytes int) (stateHash string, txs [][]byte, err error) {
	prevBlockByzantineVals, err := u.prevBlockByzantineValidators()
	if err != nil {
		return "", nil, err
	}

	// begin block lifecycle phase
	if err := u.beginBlock(prevBlockByzantineVals); err != nil {
		return "", nil, err
	}
	txs = make([][]byte, 0)
	txsTotalBz := 0
	txIdx := 0

	mempool := u.GetBus().GetUtilityModule().GetMempool()
	for !mempool.IsEmpty() {
		// NB: In order for transactions to have entered the mempool, `HandleTransaction` must have
		// been called which handles basic checks & validation.
		txBz, err := mempool.PopTx()
		if err != nil {
			return "", nil, err
		}

		tx, err := coreTypes.TxFromBytes(txBz)
		if err != nil {
			return "", nil, err
		}

		txBzSize := len(txBz)
		txsTotalBz += txBzSize

		// Exceeding maximum transaction bytes to be added in this block
		if txsTotalBz >= maxTransactionBytes {
			// Add back popped tx to be applied in a future block
			if err := mempool.AddTx(txBz); err != nil {
				return "", nil, err
			}
			break // we've reached our max
		}

		txResult, err := u.hydrateTxResult(tx, txIdx)
		if err != nil {
			u.logger.Err(err).Msg("Error in ApplyTransaction")
			// TODO(#327): Properly implement 'unhappy path' for save points
			if err := u.revertLastSavePoint(); err != nil {
				return "", nil, err
			}
			txsTotalBz -= txBzSize
			continue
		}

		// Index the transaction
		if err := u.persistenceRWContext.IndexTransaction(txResult); err != nil {
			u.logger.Fatal().Err(err).Msgf("TODO(#327): The transaction can by hydrated but not indexed. Crash the process for now: %v\n", err)
		}

		txs = append(txs, txBz)
		txIdx++
	}

	if err := u.endBlock(proposer); err != nil {
		return "", nil, err
	}

	// TODO: @deblasis - this should be from a ReadContext (the ephemeral/staging one)
	// Compute & return the new state hash
	stateHash, err = u.persistenceRWContext.ComputeStateHash()
	if err != nil {
		u.logger.Fatal().Err(err).Msg("Updating the app hash failed. TODO: Look into roll-backing the entire commit...")
	}
	u.logger.Info().Str("state_hash", stateHash).Msgf("CreateAndApplyProposalBlock finished successfully")

	return stateHash, txs, err
}

// CLEANUP: code re-use ApplyBlock() for CreateAndApplyBlock()
func (u *baseUtilityUnitOfWork) ApplyBlock() (stateHash string, txs [][]byte, err error) {
	lastByzantineValidators, err := u.prevBlockByzantineValidators()
	if err != nil {
		return "", nil, err
	}

	// begin block lifecycle phase
	if err := u.beginBlock(lastByzantineValidators); err != nil {
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
	// TODO: @deblasis - this should be from a ReadContext (the ephemeral/staging one)
	// return the app hash (consensus module will get the validator set directly)
	stateHash, err = u.persistenceRWContext.ComputeStateHash()
	if err != nil {
		u.logger.Fatal().Err(err).Msg("Updating the app hash failed. TODO: Look into roll-backing the entire commit...")
		return "", nil, utilTypes.ErrAppHash(err)
	}
	u.logger.Info().Msgf("ApplyBlock - computed state hash: %s", stateHash)

	// return the app hash; consensus module will get the validator set directly
	return stateHash, nil, nil
}

func (uow *baseUtilityUnitOfWork) Commit(quorumCert []byte) error {
	// TODO: @deblasis - change tracking here

	uow.logger.Debug().Msg("committing the rwPersistenceContext...")
	if err := uow.persistenceRWContext.Commit(uow.proposalProposerAddr, quorumCert); err != nil {
		return err
	}
	uow.persistenceRWContext = nil
	return nil
}

func (uow *baseUtilityUnitOfWork) Release() error {
	// TODO: @deblasis - change tracking reset here

	if uow.persistenceRWContext == nil {
		return nil
	}

	if err := uow.persistenceRWContext.Release(); err != nil {
		return err
	}
	uow.persistenceRWContext = nil
	return nil
}
