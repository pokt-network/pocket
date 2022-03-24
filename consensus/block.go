package consensus

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"unsafe"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

// TODO(olshansky): Sync with Andrew on the type of validation we need here.
func (m *consensusModule) validateBlock(block *types_consensus.BlockConsensusTemp) error {
	if block == nil {
		return types_consensus.ErrNilBlock
	}
	return nil
}

// This is a helper function intended to be called by a leader/validator during a view change
func (m *consensusModule) prepareBlock() (*types_consensus.BlockConsensusTemp, error) {
	if m.isReplica() {
		return nil, types_consensus.ErrReplicaPrepareBlock
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

	blockHeader := &types_consensus.BlockHeaderConsensusTemp{
		Height:            int64(m.Height),
		Hash:              appHash,
		NumTxs:            uint32(len(txs)),
		LastBlockHash:     types.GetTestState(nil).AppHash, // testing temporary
		ProposerAddress:   m.privateKey.Address(),
		QuorumCertificate: nil,
	}

	block := &types_consensus.BlockConsensusTemp{
		BlockHeader:  blockHeader,
		Transactions: txs,
	}

	return block, nil
}

// This is a helper function intended to be called by a replica/voter during a view change
func (m *consensusModule) applyBlock(block *types_consensus.BlockConsensusTemp) error {
	if m.isLeader() {
		return types_consensus.ErrLeaderApplyBLock
	}

	// TODO(olshansky): Add unit tests to verify this.
	if unsafe.Sizeof(*block) > uintptr(m.consCfg.MaxBlockBytes) {
		// TODO(olshansky) use error functions to pass params
		return fmt.Errorf("%s: %d bytes VS max of %d bytes", types_consensus.ErrBlockSizeTooLarge, unsafe.Sizeof(*block), m.consCfg.MaxBlockBytes)
	}

	if err := m.updateUtilityContext(); err != nil {
		return err
	}

	appHash, err := m.utilityContext.ApplyBlock(int64(m.Height), m.privateKey.Address(), block.Transactions, lastByzValidators)
	if err != nil {
		return err
	}

	if !bytes.Equal(block.BlockHeader.Hash, appHash) { // TODO(olshansky) blockhash is not the appHash. Discuss offline with Andrew
		return fmt.Errorf("%s: %s != %s",
			types_consensus.ErrInvalidApphash, hex.EncodeToString(block.BlockHeader.Hash), hex.EncodeToString(appHash))
	}

	return nil
}

// Creates a new Utility context and clears/nullifies any previous contexts if they exist
func (m *consensusModule) updateUtilityContext() error {
	if m.utilityContext != nil {
		m.nodeLog("[WARN] Why is the node utility context not nil when preparing a new block? Releasing for now...")
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

func (m *consensusModule) commitBlock(block *types_consensus.BlockConsensusTemp) error {
	m.nodeLog(fmt.Sprintf("🧱🧱🧱 Committing block at height %d with %d transactions 🧱🧱🧱", m.Height, len(block.Transactions)))

	if err := m.utilityContext.GetPersistanceContext().Commit(); err != nil {
		return err
	}
	m.utilityContext.ReleaseContext()
	m.utilityContext = nil

	state := types.GetTestState(nil)
	state.UpdateAppHash(hex.EncodeToString(block.BlockHeader.Hash))
	state.UpdateBlockHeight(uint64(block.BlockHeader.Height))

	return nil
}
