package unit_of_work

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/mempool"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ modules.UtilityUnitOfWork       = &leaderUtilityUnitOfWork{}
	_ modules.LeaderUtilityUnitOfWork = &leaderUtilityUnitOfWork{}
)

type leaderUtilityUnitOfWork struct {
	baseUtilityUnitOfWork
}

func NewLeaderUOW(height int64, readContext modules.PersistenceReadContext, rwPersistenceContext modules.PersistenceRWContext) *leaderUtilityUnitOfWork {
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
	log := uow.logger.With().Fields(map[string]interface{}{
		"proposer":   hex.EncodeToString(proposer),
		"maxTxBytes": maxTxBytes,
		"source":     "CreateProposalBlock",
	}).Logger()
	log.Debug().Msg("calling beginBlock")
	// begin block lifecycle phase
	if err := uow.beginBlock(); err != nil {
		return "", nil, err
	}

	log.Debug().Msg("reaping the mempool")
	txMempool := uow.GetBus().GetUtilityModule().GetMempool()
	if txs, err = uow.reapMempool(txMempool, maxTxBytes); err != nil {
		return "", nil, err
	}

	// end block lifecycle phase
	log.Debug().Msg("calling endBlock")
	if err := uow.endBlock(proposer); err != nil {
		return "", nil, err
	}

	log.Debug().Msg("computing state hash")
	// Compute & return the new state hash
	stateHash, err = uow.persistenceRWContext.ComputeStateHash()
	if err != nil {
		log.Fatal().Err(err).Bool("TODO", true).Msg("Updating the app hash failed. TODO: Look into roll-backing the entire commit...")
	}
	log.Info().Str("state_hash", stateHash).Msg("Finished successfully")

	return stateHash, txs, err
}

// reapMempool reaps transactions from the mempool up to the maximum transaction bytes allowed in a block.
func (uow *leaderUtilityUnitOfWork) reapMempool(txMempool mempool.TXMempool, maxTxBytes uint64) (txs [][]byte, err error) {
	txs = make([][]byte, 0)
	txsTotalBz := uint64(0)
	txIdx := 0
	for !txMempool.IsEmpty() {
		// NB: In order for transactions to have entered the mempool, `HandleTransaction` must have
		// been called which handles basic checks & validation.
		txBz, err := txMempool.PopTx()
		if err != nil {
			return nil, err
		}

		tx, err := coreTypes.TxFromBytes(txBz)
		if err != nil {
			return nil, err
		}

		txBzSize := uint64(len(txBz))
		txsTotalBz += txBzSize

		// Exceeding maximum transaction bytes to be added in this block
		if txsTotalBz >= maxTxBytes {
			// Add back popped tx to be applied in a future block
			if err := txMempool.AddTx(txBz); err != nil {
				return nil, err
			}
			break // we've reached our max
		}

		txResult, err := uow.hydrateTxResult(tx, txIdx)
		if err != nil {
			uow.logger.Err(err).Msg("Error in ApplyTransaction")
			// TODO(#327): Properly implement 'unhappy path' for save points
			if err := uow.revertLastSavePoint(); err != nil {
				return nil, err
			}
			txsTotalBz -= txBzSize
			continue
		}

		// TODO(#564): make sure that indexing is reversible in case of a rollback
		// Index the transaction
		if err := uow.persistenceRWContext.IndexTransaction(txResult); err != nil {
			uow.logger.Fatal().Bool("TODO", true).Err(err).Msg("TODO(#327): The transaction can by hydrated but not indexed. Crash the process for now")
		}

		txs = append(txs, txBz)
		txIdx++
	}
	return
}
