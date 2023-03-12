package unit_of_work

import (
	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
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

func (uow *leaderUtilityUnitOfWork) CreateAndApplyProposalBlock(proposer []byte, maxTxBytes uint64) (stateHash string, txs [][]byte, err error) {
	prevBlockByzantineVals, err := uow.prevBlockByzantineValidators()
	if err != nil {
		return "", nil, err
	}

	// begin block lifecycle phase
	if err := uow.beginBlock(prevBlockByzantineVals); err != nil {
		return "", nil, err
	}
	txs = make([][]byte, 0)
	txsTotalBz := uint64(0)
	txIdx := 0

	mempool := uow.GetBus().GetUtilityModule().GetMempool()
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

		txBzSize := uint64(len(txBz))
		txsTotalBz += txBzSize

		// Exceeding maximum transaction bytes to be added in this block
		if txsTotalBz >= maxTxBytes {
			// Add back popped tx to be applied in a future block
			if err := mempool.AddTx(txBz); err != nil {
				return "", nil, err
			}
			break // we've reached our max
		}

		txResult, err := uow.hydrateTxResult(tx, txIdx)
		if err != nil {
			uow.logger.Err(err).Msg("Error in ApplyTransaction")
			// TODO(#327): Properly implement 'unhappy path' for save points
			if err := uow.revertLastSavePoint(); err != nil {
				return "", nil, err
			}
			txsTotalBz -= txBzSize
			continue
		}

		// Index the transaction
		if err := uow.persistenceRWContext.IndexTransaction(txResult); err != nil {
			uow.logger.Fatal().Err(err).Msgf("TODO(#327): The transaction can by hydrated but not indexed. Crash the process for now: %v\n", err)
		}

		txs = append(txs, txBz)
		txIdx++
	}

	if err := uow.endBlock(proposer); err != nil {
		return "", nil, err
	}

	// TODO: @deblasis - this should be from a ReadContext (the ephemeral/staging one)
	// Compute & return the new state hash
	stateHash, err = uow.persistenceRWContext.ComputeStateHash()
	if err != nil {
		uow.logger.Fatal().Err(err).Msg("Updating the app hash failed. TODO: Look into roll-backing the entire commit...")
	}
	uow.logger.Info().Str("state_hash", stateHash).Msgf("CreateAndApplyProposalBlock finished successfully")

	return stateHash, txs, err
}
