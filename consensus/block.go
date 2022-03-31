package consensus

import (
	"encoding/hex"
	"unsafe"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

// TODO(olshansky): Sync with Andrew on the type of validation we need here.
func (m *consensusModule) validateBlock(block *types.Block) error {
	if block == nil {
		return typesCons.ErrNilBlock
	}
	return nil
}

// This is a helper function intended to be called by a leader/validator during a view change
func (m *consensusModule) prepareBlock() (*types.Block, error) {
	if m.isReplica() {
		return nil, typesCons.ErrReplicaPrepareBlock
	}

	if err := m.updateUtilityContext(); err != nil {
		return nil, err
	}

	txs, err := m.utilityContext.GetTransactionsForProposal(m.privateKey.Address(), maxTxBytes, lastByzValidators)
	if err != nil {
		return nil, err
	}

	appHash, err := m.utilityContext.ApplyBlock(int64(m.Height), m.privateKey.Address(), txs, lastByzValidators)
	if err != nil {
		return nil, err
	}

	blockHeader := &types.BlockHeader{
		Height:            int64(m.Height),
		Hash:              hex.EncodeToString(appHash),
		NumTxs:            uint32(len(txs)),
		LastBlockHash:     types.GetTestState(nil).AppHash, // testing temporary
		ProposerAddress:   m.privateKey.Address(),
		QuorumCertificate: nil,
	}

	block := &types.Block{
		BlockHeader:  blockHeader,
		Transactions: txs,
	}

	return block, nil
}

// This is a helper function intended to be called by a replica/voter during a view change
func (m *consensusModule) applyBlock(block *types.Block) error {
	if m.isLeader() {
		return typesCons.ErrLeaderApplyBLock
	}

	// TODO(olshansky): Add unit tests to verify this.
	if unsafe.Sizeof(*block) > uintptr(m.consCfg.MaxBlockBytes) {
		return typesCons.ErrInvalidBlockSize(uint64(unsafe.Sizeof(*block)), m.consCfg.MaxBlockBytes)
	}

	if err := m.updateUtilityContext(); err != nil {
		return err
	}

	appHash, err := m.utilityContext.ApplyBlock(int64(m.Height), m.privateKey.Address(), block.Transactions, lastByzValidators)
	if err != nil {
		return err
	}

	// TODO(olshansky) blockhash is not the appHash. Discuss offline with Andrew
	if block.BlockHeader.Hash != hex.EncodeToString(appHash) {
		return typesCons.ErrInvalidAppHash(block.BlockHeader.Hash, hex.EncodeToString(appHash))
	}

	return nil
}

// Creates a new Utility context and clears/nullifies any previous contexts if they exist
func (m *consensusModule) updateUtilityContext() error {
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

func (m *consensusModule) commitBlock(block *types.Block) error {
	m.nodeLog(typesCons.CommittingBlock(m.Height, len(block.Transactions)))

	if err := m.utilityContext.GetPersistenceContext().Commit(); err != nil {
		return err
	}
	m.utilityContext.ReleaseContext()
	m.utilityContext = nil

	state := types.GetTestState(nil)
	state.UpdateAppHash(block.BlockHeader.Hash)
	state.UpdateBlockHeight(uint64(block.BlockHeader.Height))

	return nil
}
