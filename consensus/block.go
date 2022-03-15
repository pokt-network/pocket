package consensus

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"unsafe"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

// TODO(integration): Temporary vars for integration purposes.
var (
	maxTxBytes        = 90000             // TODO(olshansky): Move this to config.json.
	lastByzValidators = make([][]byte, 0) // TODO(olshansky): Retrieve this from persistence
)

// TODO(olshansky): Sync with Andrew on the type of validation we need here.
func (m *consensusModule) isValidBlock(block *types_consensus.BlockConsensusTemp) (bool, string) {
	if block == nil {
		return false, "block is nil"
	}
	return true, "block is valid"
}

// This is a helper function intended to be called by a leader/validator during a view change
func (m *consensusModule) prepareBlock() (*types_consensus.BlockConsensusTemp, error) {
	if !m.isLeader() {
		return nil, fmt.Errorf("node should not call `prepareBlock` if it is not a leader")
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
		LastBlockHash:     types.GetTestState(nil).AppHash,
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
		return fmt.Errorf("node should not call `applyBlock` if it is not a leader")
	}

	// TODO(olshansky): Add unit tests to verify this.
	if unsafe.Sizeof(*block) > uintptr(m.consCfg.MaxBlockBytes) {
		return fmt.Errorf("block size is too large: %d bytes VS max of %d bytes", unsafe.Sizeof(*block), m.consCfg.MaxBlockBytes)
	}

	if err := m.updateUtilityContext(); err != nil {
		return err
	}

	appHash, err := m.utilityContext.ApplyBlock(int64(m.Height), m.privateKey.Address(), block.Transactions, lastByzValidators)
	if err != nil {
		return err
	}

	if !bytes.Equal(block.BlockHeader.Hash, appHash) {
		return fmt.Errorf("block hash being applied does not equal that from utility: %s != %s", hex.EncodeToString(block.BlockHeader.Hash), hex.EncodeToString(appHash))
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
