package consensus

import (
	"encoding/hex"
	"fmt"
	"strconv"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

func (m *consensusModule) prepareBlock() (*types_consensus.BlockConsensusTemp, error) {
	state := types.GetTestState(nil)

	if m.utilityContext != nil {
		m.nodeLog("[WARN] Why is the node utility context not nil when preparing a new block? Realising for now...")
		m.utilityContext.ReleaseContext()
	}

	utilContext, err := m.GetBus().GetUtilityModule().NewContext(int64(m.Height))
	if err != nil {
		return nil, err
	}
	m.utilityContext = utilContext

	maxTxBytes := 90000                    // TODO(olshansky): Retrieve this from global configs
	lastByzValidators := make([][]byte, 0) // TODO(olshansky): Retrieve this from persistence
	txs, err := m.utilityContext.GetTransactionsForProposal(state.PrivateKey.Address(), maxTxBytes, lastByzValidators)
	if err != nil {
		return nil, err
	}

	header := &types_consensus.BlockHeaderConsensusTemp{
		Height:            int64(m.Height),
		Hash:              strconv.Itoa(int(m.Height)),
		Time:              nil, // TODO(olshansky): What should this be?
		NumTxs:            uint32(len(txs)),
		LastBlockHash:     state.AppHash,
		ProposerAddress:   state.PrivateKey.Address(),
		QuorumCertificate: nil, // TODO(olshansky): See the comment in `block_cons_temp.proto`
	}

	block := &types_consensus.BlockConsensusTemp{
		BlockHeader:  header,
		Transactions: txs,
	}

	return block, nil
}

// TODO(olshansky): Implement this properly....
func (m *consensusModule) isValidBlock(block *types_consensus.BlockConsensusTemp) (bool, string) {
	if block == nil {
		return false, "Block is nil"
	}
	return true, ""

}

// TODO: Should this be async?
func (m *consensusModule) deliverTxToUtility(block *types_consensus.BlockConsensusTemp) error {
	utilityModule := m.GetBus().GetUtilityModule()
	m.utilityContext, _ = utilityModule.NewContext(int64(m.Height))
	proposer := []byte(strconv.Itoa(int(m.NodeId)))
	lastByzValidators := make([][]byte, 0) // INTEGRATION_TEMP: m.utilityContext.GetPersistanceContext().GetLastByzValidators

	appHash, err := m.utilityContext.ApplyBlock(int64(m.Height), proposer, block.Transactions, lastByzValidators)
	if err != nil {
		return err
	}

	// INTEGRATION_TEMP: Make sure the BlockHeader uses the same encoding as the appHash
	if block.BlockHeader.Hash != hex.EncodeToString(appHash) {
		return fmt.Errorf("[ERROR] Why is the block header hash not what utility returned?")
	}

	return nil
}

func (m *consensusModule) commitBlock(block *types_consensus.BlockConsensusTemp) error {
	m.nodeLog(fmt.Sprintf("COMMITTING BLOCK AT HEIGHT %d. WITH TRANSACTION COUNT: %d", m.Height, len(block.Transactions)))
	if err := m.utilityContext.GetPersistanceContext().Commit(); err != nil {
		return err
	}
	m.utilityContext.ReleaseContext()
	m.utilityContext = nil

	//utilityModule := m.GetBus().GetUtilityModule()
	//if err := utilityModule.EndBlock(nil); err != nil {
	//	m.paceMaker.InterruptRound()
	//	return err
	//}

	pocketState := types.GetTestState(nil)
	pocketState.UpdateAppHash(block.BlockHeader.Hash)
	pocketState.UpdateBlockHeight(uint64(block.BlockHeader.Height))

	// TODO: Something with the persistence module?
	// persistenceModule := m.GetBus().GetpersistenceModule()

	return nil
}
