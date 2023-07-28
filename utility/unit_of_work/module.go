package unit_of_work

import (
	"errors"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

const (
	leaderUtilityUOWModuleName  = "leader_utility_UOW"
	replicaUtilityUOWModuleName = "replica_utility_UOW"
)

var _ modules.UtilityUnitOfWork = &baseUtilityUnitOfWork{}

type baseUtilityUnitOfWork struct {
	base_modules.IntegrableModule

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

	stateHash string
}

func (uow *baseUtilityUnitOfWork) SetProposalBlock(blockHash string, proposerAddr []byte, txs [][]byte) error {
	uow.proposalStateHash = blockHash
	uow.proposalProposerAddr = proposerAddr
	uow.proposalBlockTxs = txs
	return nil
}

// ApplyBlock atomically applies a block to the persistence layer for a given height.
func (uow *baseUtilityUnitOfWork) ApplyBlock() error {
	log := uow.logger.With().Fields(map[string]interface{}{
		"source": "ApplyBlock",
	}).Logger()

	log.Debug().Msg("checking if proposal block has been set")
	if !uow.isProposalBlockSet() {
		return coreTypes.ErrProposalBlockNotSet()
	}

	// initialize a new savepoint before applying the block
	if err := uow.newSavePoint(); err != nil {
		return err
	}

	// begin block lifecycle phase
	log.Debug().Msg("calling beginBlock")
	if err := uow.beginBlock(); err != nil {
		return err
	}

	// processProposalBlockTransactions indexes the transactions into the TxIndexer.
	// If it fails, it returns an error which triggers a rollback below to undo the changes
	// that processProposalBlockTransactions could have caused.
	log.Debug().Msg("processing transactions from proposal block")
	if err := uow.processProposalBlockTransactions(); err != nil {
		rollErr := uow.revertToLastSavepoint()
		return errors.Join(rollErr, err)
	}

	// end block lifecycle phase calls endBlock and reverts to the last known savepoint if it encounters any errors
	log.Debug().Msg("calling endBlock")
	if err := uow.endBlock(uow.proposalProposerAddr); err != nil {
		rollErr := uow.revertToLastSavepoint()
		return errors.Join(rollErr, err)
	}

	// return the app hash (consensus module will get the validator set directly)
	stateHash, err := uow.persistenceRWContext.ComputeStateHash()
	if err != nil {
		rollErr := uow.persistenceRWContext.RollbackToSavePoint()
		return coreTypes.ErrAppHash(errors.Join(err, rollErr))
	}

	// IMPROVE(#655): this acts as a feature flag to allow tests to ignore the check if needed, ideally the tests should have a way to determine
	// the hash and set it into the proposal block it's currently hard to do because the state is different at every test run (non-determinism)
	if uow.proposalStateHash != IgnoreProposalBlockCheckHash {
		if uow.proposalStateHash != stateHash {
			return uow.revertToLastSavepoint()
		}
	}

	log.Info().Str("state_hash", stateHash).Msgf("🧱 ApplyBlock succeeded!")

	uow.stateHash = stateHash

	return nil
}

func (uow *baseUtilityUnitOfWork) Commit(quorumCert []byte) error {
	uow.logger.Debug().Msg("committing the rwPersistenceContext...")
	if err := uow.persistenceRWContext.Commit(uow.proposalProposerAddr, quorumCert); err != nil {
		return err
	}
	uow.persistenceRWContext = nil
	return nil
}

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

// isProposalBlockSet returns true if the proposal block has been set.
// TODO: it should also check that uow.proposalBlockTxs is not empty but if we do, tests fail.
func (uow *baseUtilityUnitOfWork) isProposalBlockSet() bool {
	return uow.proposalStateHash != "" && uow.proposalProposerAddr != nil
}

// processProposalBlockTransactions processes the transactions from the proposal block stored in the current
// unit of work. It applies the transactions to the persistence context, indexes them, and removes that from
// the mempool if they are present.
func (uow *baseUtilityUnitOfWork) processProposalBlockTransactions() (err error) {
	// CONSIDERATION: should we check that `uow.proposalBlockTxs` is not nil and return an error if so or allow empty blocks?
	// For reference, see Tendermint: https://docs.tendermint.com/v0.34/tendermint-core/configuration.html#empty-blocks-vs-no-empty-blocks
	txMempool := uow.GetBus().GetUtilityModule().GetMempool()
	for index, txProtoBytes := range uow.proposalBlockTxs {
		tx, err := coreTypes.TxFromBytes(txProtoBytes)
		if err != nil {
			return err
		}

		txHash, err := tx.Hash()
		if err != nil {
			return err
		}

		if uow.logger.GetLevel().String() == "debug" {
			uow.logger.Debug().Str("tx", txHash).Msgf("processing transaction: %+v", tx)
		}

		if err := tx.ValidateBasic(); err != nil {
			return err
		}

		idxTx, err := uow.HandleTransaction(tx, index)
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
		if err := uow.persistenceRWContext.IndexTransaction(idxTx); err != nil {
			uow.logger.Fatal().Err(err).Msg("TODO(#327): We can apply the transaction but not index it. Crash the process for now")
		}
	}
	return nil
}

// GetStateHash returns the state hash of the unit of work. It is only available after the block has been applied.
func (uow *baseUtilityUnitOfWork) GetStateHash() string {
	return uow.stateHash
}
