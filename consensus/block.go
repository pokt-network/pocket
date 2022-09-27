package consensus

import (
	"log"
	"unsafe"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
)

func (m *ConsensusModule) commitBlock(block *typesCons.Block) error {
	m.nodeLog(typesCons.CommittingBlock(m.Height, len(block.Transactions)))

	// Store the block in the KV store
	codec := codec.GetCodec()
	blockProtoBytes, err := codec.Marshal(block)
	if err != nil {
		return err
	}

	// IMPROVE(olshansky): temporary solution. `ApplyBlock` above applies the
	// transactions to the postgres database, and this stores it in the KV store upon commitment.
	// Instead of calling this directly, an alternative solution is to store the block metadata in
	// the persistence context and have `CommitPersistenceContext` do this under the hood. However,
	// additional `Block` metadata will need to be passed through and may change when we merkle the
	// state hash.
	if err := m.UtilityContext.StoreBlock(blockProtoBytes); err != nil {
		return err
	}

	// Commit the utility context
	if err := m.UtilityContext.CommitPersistenceContext(); err != nil {
		return err
	}
	// Release the utility context
	m.UtilityContext.ReleaseContext()
	m.UtilityContext = nil
	// Update the last app hash
	m.lastAppHash = block.BlockHeader.Hash

	return nil
}

// TODO: Add unit tests specific to block validation
func (m *ConsensusModule) validateBlockBasic(block *typesCons.Block) error {
	if block == nil && m.Step != NewRound {
		return typesCons.ErrNilBlock
	}

	if block != nil && m.Step == NewRound {
		return typesCons.ErrBlockExists
	}

	if block != nil && unsafe.Sizeof(*block) > uintptr(m.consGenesis.MaxBlockBytes) {
		return typesCons.ErrInvalidBlockSize(uint64(unsafe.Sizeof(*block)), m.consGenesis.MaxBlockBytes)
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
func (m *ConsensusModule) refreshUtilityContext() error {
	// Catch-all structure to release the previous utility context if it wasn't properly cleaned up.
	// Ideally, this should not be called.
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
