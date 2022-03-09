package consensus

import (
	"bytes"
	"encoding/hex"
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

// TODO(olshansky): Implement this properly....
func (m *consensusModule) isValidBlock(block *types_consensus.BlockConsensusTemp) (bool, string) {
	if block == nil {
		return false, "block is nil"
	}
	return true, "block is valid"
}

func (m *consensusModule) prepareBlock() (*types_consensus.BlockConsensusTemp, error) {
	if err := m.updateUtilityContext(); err != nil {
		return nil, err
	}

	maxTxBytes := 90000                    // TODO(olshansky): Move this to config.json.
	lastByzValidators := make([][]byte, 0) // TODO(olshansky): Retrieve this from persistence
	txs, err := m.utilityContext.GetTransactionsForProposal(m.privateKey.Address(), maxTxBytes, lastByzValidators)
	if err != nil {
		return nil, err
	}

	height := int64(m.Height)
	proposer := m.privateKey.Address()
	appHash, err := m.utilityContext.ApplyBlock(height, proposer, txs, lastByzValidators)
	if err != nil {
		return nil, err
	}

	header := &types_consensus.BlockHeaderConsensusTemp{
		Height:            height,
		Hash:              appHash,
		NumTxs:            uint32(len(txs)),
		LastBlockHash:     types.GetTestState(nil).AppHash,
		ProposerAddress:   m.privateKey.Address(),
		QuorumCertificate: nil,
	}

	block := &types_consensus.BlockConsensusTemp{
		BlockHeader:  header,
		Transactions: txs,
	}

	return block, nil
}

func (m *consensusModule) applyBlock(block *types_consensus.BlockConsensusTemp) error {
	if err := m.updateUtilityContext(); err != nil {
		return err
	}

	lastByzValidators := make([][]byte, 0) // TODO(olshansky): Retrieve this from persistence
	height := int64(m.Height)
	proposer := m.privateKey.Address()

	appHash, err := m.utilityContext.ApplyBlock(height, proposer, block.Transactions, lastByzValidators)
	if err != nil {
		return err
	}
	if !bytes.Equal(block.BlockHeader.Hash, appHash) {
		return fmt.Errorf("block hash does not match app hash: %s != %s", hex.EncodeToString(block.BlockHeader.Hash), hex.EncodeToString(appHash))
	}

	return nil
}

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
	m.nodeLog(fmt.Sprintf("COMMITTING BLOCK AT HEIGHT %d WITH %d TRANSACTIONS", m.Height, len(block.Transactions)))

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
