package consensus

import (
	"log"
	"unsafe"

	typesCons "github.com/pokt-network/pocket/internal/consensus/types"
)

func (m *consensusModule) commitBlock(block *typesCons.Block) error {
	// Commit the context
	if err := m.utilityContext.Commit(block.BlockHeader.QuorumCertificate); err != nil {
		return err
	}
	m.nodeLog(typesCons.CommittingBlock(m.height, len(block.Transactions)))

	// Release the context
	if err := m.utilityContext.Release(); err != nil {
		log.Println("[WARN] Error releasing utility context: ", err)
	}

	m.utilityContext = nil

	return nil
}

// TODO: Add unit tests specific to block validation
// IMPROVE: (olshansky) rename to provide clarity of operation. ValidateBasic() is typically a stateless check not stateful
func (m *consensusModule) validateBlockBasic(block *typesCons.Block) error {
	if block == nil && m.step != NewRound {
		return typesCons.ErrNilBlock
	}

	if block != nil && m.step == NewRound {
		return typesCons.ErrBlockExists
	}

	if block != nil && unsafe.Sizeof(*block) > uintptr(m.consGenesis.GetMaxBlockBytes()) {
		return typesCons.ErrInvalidBlockSize(uint64(unsafe.Sizeof(*block)), m.consGenesis.GetMaxBlockBytes())
	}

	// If the current block being processed (i.e. voted on) by consensus is non nil, we need to make
	// sure that the data (height, round, step, txs, etc) is the same before we start validating the signatures
	if m.block != nil {
		// DISCUSS: The only difference between blocks from one step to another is the QC, so we need
		//          to determine where/how to validate this
		if protoHash(m.block) != protoHash(block) {
			log.Println("[TECHDEBT][ERROR] The block being processed is not the same as that received by the consensus module ")
		}
	}

	return nil
}

// Creates a new Utility context and clears/nullifies any previous contexts if they exist
func (m *consensusModule) refreshUtilityContext() error {
	// Catch-all structure to release the previous utility context if it wasn't properly cleaned up.
	// Ideally, this should not be called.
	if m.utilityContext != nil {
		m.nodeLog(typesCons.NilUtilityContextWarning)
		if err := m.utilityContext.Release(); err != nil {
			log.Printf("[WARN] Error releasing utility context: %v\n", err)
		}
		m.utilityContext = nil
	}

	// Only one write context can exist at a time, and the utility context needs to instantiate
	// a new one to modify the state.
	if err := m.GetBus().GetPersistenceModule().ReleaseWriteContext(); err != nil {
		log.Printf("[WARN] Error releasing persistence write context: %v\n", err)
	}

	utilityContext, err := m.GetBus().GetUtilityModule().NewContext(int64(m.height))
	if err != nil {
		return err
	}

	m.utilityContext = utilityContext
	return nil
}
