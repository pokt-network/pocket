package consensus

import (
	"log"
	"unsafe"

	typesCons "github.com/pokt-network/pocket/consensus/types"
)

func (m *consensusModule) commitBlock(block *typesCons.Block) error {
	m.nodeLog(typesCons.CommittingBlock(m.Height, len(block.Transactions)))

	// Commit and release the context
	if err := m.utilityContext.Commit(block.BlockHeader.QuorumCertificate); err != nil {
		return err
	}

	if err := m.utilityContext.Release(); err != nil {
		return err
	}
	m.utilityContext = nil

	m.lastAppHash = block.BlockHeader.Hash

	return nil
}

// TODO: Add unit tests specific to block validation
// IMPROVE: (olshansky) rename to provide clarity of operation. ValidateBasic() is typically a stateless check not stateful
func (m *consensusModule) validateBlockBasic(block *typesCons.Block) error {
	if block == nil && m.Step != NewRound {
		return typesCons.ErrNilBlock
	}

	if block != nil && m.Step == NewRound {
		return typesCons.ErrBlockExists
	}

	if block != nil && unsafe.Sizeof(*block) > uintptr(m.consGenesis.GetMaxBlockBytes()) {
		return typesCons.ErrInvalidBlockSize(uint64(unsafe.Sizeof(*block)), m.consGenesis.GetMaxBlockBytes())
	}

	// If the current block being processed (i.e. voted on) by consensus is non nil, we need to make
	// sure that the data (height, round, step, txs, etc) is the same before we start validating the signatures
	if m.Block != nil {
		// DISCUSS: The only difference between blocks from one step to another is the QC, so we need
		//          to determine where/how to validate this
		if protoHash(m.Block) != protoHash(block) {
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
		err := m.utilityContext.Release()
		m.utilityContext = nil
		if err != nil {
			return err
		}
	}

	utilityContext, err := m.GetBus().GetUtilityModule().NewContext(int64(m.Height))
	if err != nil {
		return err
	}

	m.utilityContext = utilityContext
	return nil
}
