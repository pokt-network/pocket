package consensus

import (
	"fmt"
	"unsafe"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

func (m *consensusModule) commitBlock(block *coreTypes.Block) error {
	utilityUnitOfWork := m.utilityUnitOfWork
	if utilityUnitOfWork == nil {
		return fmt.Errorf("utility uow is nil")
	}

	// Commit & release the unit of work
	if err := utilityUnitOfWork.Commit(block.BlockHeader.QuorumCertificate); err != nil {
		return err
	}
	if err := utilityUnitOfWork.Release(); err != nil {
		m.logger.Warn().Err(err).Msg("failed to release utility unit of work after commit")
	}
	m.utilityUnitOfWork = nil

	m.logger.Info().
		Fields(
			map[string]any{
				"height":       block.BlockHeader.Height,
				"transactions": len(block.Transactions),
			}).
		Msg("🧱🧱🧱 Committing block 🧱🧱🧱")

	return nil
}

// isBlockInMessageValidBasic does basic validation of the block in the hotstuff message such as:
// - validating if the block could/should be nil
// - the state hash of the block
// - the size of the block
// ADDTEST: Add unit tests specific to block validation
func (m *consensusModule) isBlockInMessageValidBasic(msg *typesCons.HotstuffMessage) (bool, error) {
	block := msg.GetBlock()
	step := msg.GetStep()

	if block == nil {
		if step != NewRound {
			return false, fmt.Errorf("validateBlockBasic failed - block is nil during step %s", typesCons.StepToString[m.step])
		}
		m.logger.Debug().Msg("✅ NewRound block is nil.")
		return true, nil
	}

	if block != nil && step == NewRound {
		return false, fmt.Errorf("validateBlockBasic failed - block is not nil during step %s", typesCons.StepToString[m.step])
	}

	if block != nil && unsafe.Sizeof(*block) > uintptr(m.genesisState.GetMaxBlockBytes()) {
		return false, typesCons.ErrInvalidBlockSize(uint64(unsafe.Sizeof(*block)), m.genesisState.GetMaxBlockBytes())
	}

	// If the current block being processed (i.e. voted on) by consensus is non nil, we need to make
	// sure that the data (height, round, step, txs, etc) is the same before we start validating the signatures
	if m.block != nil {
		if m.block.BlockHeader.StateHash != block.BlockHeader.StateHash {
			return false, fmt.Errorf("validateBlockBasic failed - block hash is not the same as the current block being processed by consensus")
		}

		// DISCUSS: The only difference between blocks from one step to another is the QC, so we need
		//          to determine where/how to validate this
		if protoHash(m.block) != protoHash(block) {
			m.logger.Warn().Bool("TECHDEBT", true).Msg("WalidateBlockBasic warning - block hash is the same but serialization is not")
		}
	}

	return true, nil
}

// refreshUtilityUnitOfWork is a helper that creates a new Utility Unit Of Work and clears/nullifies a previous one if it exists
func (m *consensusModule) refreshUtilityUnitOfWork() error {
	// m.m.Lock()
	// defer m.m.Unlock()

	// Catch-all structure to release the previous utility UOW if it wasn't properly cleaned up.
	utilityUnitOfWork := m.utilityUnitOfWork

	// TECHDEBT: This should, theoretically, never happen. Need to identify all
	// code paths where it does and fix it.
	if utilityUnitOfWork != nil {
		m.logger.Warn().Bool("TODO", true).Msg(typesCons.NilUtilityUOWWarning)
		if err := utilityUnitOfWork.Release(); err != nil {
			m.logger.Warn().Err(err).Msg("failed to release utility unit of work")
		}
		m.utilityUnitOfWork = nil
	}

	// Only one write context can exist at a time, and the utility unit of work needs to instantiate
	// a new one to modify the state.
	if err := m.GetBus().GetPersistenceModule().ReleaseWriteContext(); err != nil {
		m.logger.Warn().Err(err).Msg("Error releasing persistence write context")
	}

	utilityUOW, err := m.GetBus().GetUtilityModule().NewUnitOfWork(int64(m.height))
	if err != nil {
		return err
	}
	m.utilityUnitOfWork = utilityUOW

	return nil
}
