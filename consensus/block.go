package consensus

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
)

// TODO(olshansky): Sync with Andrew on the type of validation we need here.
func (m *ConsensusModule) validateBlock(block *typesCons.Block) error {
	if block == nil {
		return typesCons.ErrNilBlock
	}
	return nil
}

// Creates a new Utility context and clears/nullifies any previous contexts if they exist
func (m *ConsensusModule) refreshUtilityContext() error {
	// This is a catch-all to release the previous utility context if it wasn't cleaned up
	// in the proper lifecycle (e.g. catch up, error, network partition, etc...). Ideally, this
	// should not be called.
	if m.utilityContext != nil {
		m.nodeLog(typesCons.NilUtilityContextWarning)
		m.utilityContext.ReleaseContext()
		m.utilityContext = nil
	}

	utilityContext, err := m.GetBus().GetUtilityModule().NewContext(int64(m.Height))
	if err != nil {
		return err
	}

	m.utilityContext = utilityContext
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
	if err := m.utilityContext.CommitPersistenceContext(); err != nil {
		return err
	}

	m.utilityContext.ReleaseContext()
	m.utilityContext = nil

	m.lastAppHash = block.BlockHeader.Hash

	return nil
}
