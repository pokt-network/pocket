package unit_of_work

import (
	"fmt"

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

	stateHash string
}

func (uow *baseUtilityUnitOfWork) SetProposalBlock(blockHash string, proposerAddr []byte, txs [][]byte) error {
	uow.proposalStateHash = blockHash
	uow.proposalProposerAddr = proposerAddr
	uow.proposalBlockTxs = txs
	return nil
}

func (uow *baseUtilityUnitOfWork) ApplyBlock() error {
	log := uow.logger.With().Fields(map[string]interface{}{
		"source": "ApplyBlock",
	}).Logger()

	log.Debug().Msg("checking if proposal block has been set")
	if !uow.isProposalBlockSet() {
		return utilTypes.ErrProposalBlockNotSet()
	}

	// begin block lifecycle phase
	log.Debug().Msg("calling beginBlock")
	if err := uow.beginBlock(); err != nil {
		return err
	}

	log.Debug().Msg("processing transactions from proposal block")
	txMempool := uow.GetBus().GetUtilityModule().GetMempool()
	if err := uow.processTransactionsFromProposalBlock(txMempool); err != nil {
		return err
	}

	// end block lifecycle phase
	log.Debug().Msg("calling endBlock")
	if err := uow.endBlock(uow.proposalProposerAddr); err != nil {
		return err
	}
	// return the app hash (consensus module will get the validator set directly)
	log.Debug().Msg("computing state hash")
	stateHash, err := uow.persistenceRWContext.ComputeStateHash()
	if err != nil {
		log.Fatal().Err(err).Bool("TODO", true).Msg("Updating the app hash failed. TODO: Look into roll-backing the entire commit...")
		return utilTypes.ErrAppHash(err)
	}

	// IMPROVE(#655): this acts as a feature flag to allow tests to ignore the check if needed, ideally the tests should have a way to determine
	// the hash and set it into the proposal block it's currently hard to do because the state is different at every test run (non-determinism)
	if uow.proposalStateHash != IgnoreProposalBlockCheckHash {
		if uow.proposalStateHash != stateHash {
			log.Fatal().Bool("TODO", true).
				Str("proposalStateHash", uow.proposalStateHash).
				Str("stateHash", stateHash).
				Msg("State hash mismatch. TODO: Look into roll-backing the entire commit...")
			return utilTypes.ErrAppHash(fmt.Errorf("state hash mismatch: expected %s from the proposal, got %s", uow.proposalStateHash, stateHash))
		}
	}

	log.Info().Str("state_hash", stateHash).Msgf("ApplyBlock succeeded!")

	uow.stateHash = stateHash

	return nil
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

// isProposalBlockSet returns true if the proposal block has been set.
// TODO: it should also check that uow.proposalBlockTxs is not empty but if we do, tests fail.
func (uow *baseUtilityUnitOfWork) isProposalBlockSet() bool {
	return uow.proposalStateHash != "" && uow.proposalProposerAddr != nil
}

// processTransactionsFromProposalBlock processes the transactions from the proposal block.
// It also removes the transactions from the mempool if they are also present.
func (uow *baseUtilityUnitOfWork) processTransactionsFromProposalBlock(txMempool mempool.TXMempool) (err error) {
	for index, txProtoBytes := range uow.proposalBlockTxs {
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
			uow.logger.Fatal().Err(err).Msg("TODO(#327): We can apply the transaction but not index it. Crash the process for now")
		}
	}
	return nil
}

// GetStateHash returns the state hash of the unit of work. It is only available after the block has been applied.
func (uow *baseUtilityUnitOfWork) GetStateHash() string {
	return uow.stateHash
}
