package consensus

import (
	"encoding/hex"
	"fmt"
	"strconv"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
)

func (m *consensusModule) prepareBlock() (*types_consensus.BlockConsensusTemp, error) {
	//if m.utilityContext != nil {
	//	m.nodeLog("[WARN] Why is the node utility context not nil when preparing a new block?. Realising for now...")
	//	m.utilityContext.ReleaseContext()
	//}
	fmt.Println("creating new context for prepareBlock()")
	utilContext, err := m.GetBus().GetUtilityModule().NewContext(int64(m.Height))
	if err != nil {
		return nil, err
	}
	m.utilityContext = utilContext
	//valMap := shared.GetTestState().ValidatorMap
	maxTxBytes := 90000 // INTEGRATION_TEMP
	//proposer := []byte(strconv.Itoa(int(m.NodeId)))
	pk, _ := crypto.GeneratePrivateKey()
	lastByzValidators := make([][]byte, 0) // INTEGRATION_TEMP: m.utilityContext.GetPersistanceContext().GetLastByzValidators
	txs, err := m.utilityContext.GetTransactionsForProposal(pk.PublicKey().Address(), maxTxBytes, lastByzValidators)
	if err != nil {
		return nil, err
	}

	pocketState := types.GetTestState()

	header := &types_consensus.BlockHeaderConsensusTemp{
		Height: int64(m.Height),
		Hash:   strconv.Itoa(int(m.Height)),

		LastBlockHash: pocketState.AppHash,
		// ProposerAddress: []byte(pocketState.Address),
		// ProposerId:      uint32(m.NodeId),
		// QuorumCertificate // TODO
	}

	block := &types_consensus.BlockConsensusTemp{
		BlockHeader:  header,
		Transactions: txs,
	}

	return block, nil
}

func (m *consensusModule) isValidBlock(block *types_consensus.BlockConsensusTemp) bool {
	if block == nil {
		return false
	}
	return true

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

	pocketState := types.GetTestState()
	pocketState.UpdateAppHash(block.BlockHeader.Hash)
	pocketState.UpdateBlockHeight(uint64(block.BlockHeader.Height))

	// TODO: Something with the persistence module?
	// persistenceModule := m.GetBus().GetpersistenceModule()

	return nil
}
