package unit_of_work

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/mempool"
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
func (uow *baseUtilityUnitOfWork) ApplyBlock() (stateHash string, txs [][]byte, err error) {
	log := uow.logger.With().Fields(map[string]interface{}{
		"source": "ApplyBlock",
	}).Logger()

	log.Debug().Msg("checking if proposal block has been set")
	if !uow.isProposalBlockSet() {
		return "", nil, utilTypes.ErrProposalBlockNotSet()
	}

	// begin block lifecycle phase
	log.Debug().Msg("calling beginBlock")
	if err := uow.beginBlock(); err != nil {
		return "", nil, err
	}

	log.Debug().Msg("processing transactions from proposal block")
	txMempool := uow.GetBus().GetUtilityModule().GetMempool()
	if err := uow.processTransactionsFromProposalBlock(txMempool, uow.proposalBlockTxs); err != nil {
		return "", nil, err
	}

	// end block lifecycle phase
	log.Debug().Msg("calling endBlock")
	if err := uow.endBlock(uow.proposalProposerAddr); err != nil {
		return "", nil, err
	}
	// TODO(@deblasis): this should be from a ReadContext (the ephemeral/staging one)
	// return the app hash (consensus module will get the validator set directly)
	log.Debug().Msg("computing state hash")
	stateHash, err = uow.persistenceRWContext.ComputeStateHash()
	if err != nil {
		log.Fatal().Err(err).Msg("Updating the app hash failed. TODO: Look into roll-backing the entire commit...")
		return "", nil, utilTypes.ErrAppHash(err)
	}
	log.Info().Str("state_hash", stateHash).Msgf("ApplyBlock succeeded!")

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
	return uow.proposalStateHash != "" && uow.proposalProposerAddr != nil
}

// processTransactionsFromProposalBlock processes the transactions from the proposal block.
// It also removes the transactions from the mempool if they are also present.
func (uow *baseUtilityUnitOfWork) processTransactionsFromProposalBlock(txMempool mempool.TXMempool, txsBytes [][]byte) (err error) {
	for index, txProtoBytes := range txsBytes {
		tx, err := coreTypes.TxFromBytes(txProtoBytes)
		if err != nil {
			return err
		}
		if err := tx.ValidateBasic(); err != nil {
			return err
		}

		txResult, err := uow.hydrateTxResult(tx, index)
		if err != nil {
			return err
		}

		txHash, err := tx.Hash()
		if err != nil {
			return err
		}

		if txMempool.Contains(txHash) {
			if err := txMempool.RemoveTx(txProtoBytes); err != nil {
				return err
			}
			uow.logger.Info().Str("tx_hash", txHash).Msg("Applying tx that WAS in the local mempool")
		} else {
			uow.logger.Info().Str("tx_hash", txHash).Msg("Applying tx that WAS NOT in the local mempool")
		}

		// TODO(#564): make sure that indexing is reversible in case of a rollback
		if err := uow.persistenceRWContext.IndexTransaction(txResult); err != nil {
			uow.logger.Fatal().Err(err).Msgf("TODO(#327): We can apply the transaction but not index it. Crash the process for now: %v\n", err)
		}
	}
	return nil
}
