package consensus

import (
	"bytes"
	"encoding/hex"
	"unsafe"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
)

// TODO(olshansky): Sync with Andrew on the type of validation we need here.
func (m *consensusModule) validateBlock(block *typesCons.BlockConsensusTemp) error {
	if block == nil {
		return typesCons.ErrNilBlock
	}
	return nil
}

// This is a helper function intended to be called by a leader/validator during a view change
func (m *consensusModule) prepareBlock() (*typesCons.BlockConsensusTemp, error) {
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

	blockHeader := &typesCons.BlockHeaderConsensusTemp{
		Height:            int64(m.Height),
		Hash:              appHash,
		NumTxs:            uint32(len(txs)),
		LastBlockHash:     typesGenesis.GetNodeState(nil).AppHash, // testing temporary
		ProposerAddress:   m.privateKey.Address(),
		QuorumCertificate: nil,
	}

	block := &typesCons.BlockConsensusTemp{
		BlockHeader:  blockHeader,
		Transactions: txs,
	}

	return block, nil
}

// This is a helper function intended to be called by a replica/voter during a view change
func (m *consensusModule) applyBlock(block *typesCons.BlockConsensusTemp) error {
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
	if !bytes.Equal(block.BlockHeader.Hash, appHash) {
		return typesCons.ErrInvalidAppHash(hex.EncodeToString(block.BlockHeader.Hash), hex.EncodeToString(appHash))
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

func (m *consensusModule) commitBlock(block *typesCons.BlockConsensusTemp) error {
	m.nodeLog(typesCons.CommittingBlock(m.Height, len(block.Transactions)))

	if err := m.utilityContext.GetPersistenceContext().Commit(); err != nil {
		return err
	}
	m.utilityContext.ReleaseContext()
	m.utilityContext = nil

	state := typesGenesis.GetNodeState(nil)
	state.UpdateAppHash(hex.EncodeToString(block.BlockHeader.Hash))
	state.UpdateBlockHeight(uint64(block.BlockHeader.Height))

	return nil
}
