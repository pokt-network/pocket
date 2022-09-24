package consensus

import (
	"log"
	"unsafe"

	typesCons "github.com/pokt-network/pocket/consensus/types"
)

// TODO: Add additional basic block metadata validation w/ unit tests
func (m *ConsensusModule) validateBlockBasic(block *typesCons.Block) error {
	if block == nil {
		return typesCons.ErrNilBlock
	}

	if unsafe.Sizeof(*block) > uintptr(m.MaxBlockBytes) {
		return typesCons.ErrInvalidBlockSize(uint64(unsafe.Sizeof(*block)), m.MaxBlockBytes)
	}

	// If the current block being processed (i.e. voted on) by consensus is non nil, we need to make
	// sure that the data (height, round, step, txs, etc) is the same before we start validating the signatures
	if m.Block != nil {
		// DISCUSS: The only difference between blocks from one step to another is the QC, so we need
		// to determine where/how to validate this
		if protoHash(m.Block) != protoHash(block) {
			log.Println("[TECHDEBT][ERROR] The block being processed is not the same as that received by the consensus module ")
		}
	}

	return nil
}

// Creates a new Utility context and clears/nullifies any previous contexts if they exist
func (m *ConsensusModule) refreshUtilityContext() error {
	// This is a catch-all to release the previous utility context if it wasn't cleaned up
	// in the proper lifecycle (e.g. catch up, error, network partition, etc...). Ideally, this
	// should not be called.
	if m.UtilityContext != nil {
		m.nodeLog(typesCons.NilUtilityContextWarning)
		m.UtilityContext.ReleaseContext()
		m.UtilityContext = nil
	}

	utilityContext, err := m.GetBus().GetUtilityModule().NewContext(int64(m.Height))
	if err != nil {
		return err
	}

	m.UtilityContext = utilityContext
	return nil
}

func (m *ConsensusModule) commitBlock(block *typesCons.Block) error {
	m.nodeLog(typesCons.CommittingBlock(m.Height, len(block.Transactions)))

	// Store the block in the KV store
	// codec := codec.GetCodec()
	// blockProtoBytes, err := codec.Marshal(block)
	// if err != nil {
	// 	return err
	// }

	// Commit and release the context
	if err := m.UtilityContext.CommitPersistenceContext(); err != nil {
		return err
	}

	m.UtilityContext.ReleaseContext()
	m.UtilityContext = nil

	m.lastAppHash = block.BlockHeader.Hash

	return nil
}
