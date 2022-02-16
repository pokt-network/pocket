package consensus

import (
	"encoding/hex"
	"fmt"
	types2 "pocket/consensus/types"
	"pocket/shared/crypto"
	"pocket/shared/types"
	"strconv"
)

func (m *ConsensusModule) prepareBlock() (*types.BlockConsTemp, error) {
	if m.UtilityContext != nil {
		m.nodeLog("[WARN] Why is the node utility context not nil when preparing a new block?. Realising for now...")
		m.UtilityContext.ReleaseContext()
	}
	utilContext, err := m.GetBus().GetUtilityModule().NewContext(int64(m.Height))
	if err != nil {
		return nil, err
	}
	m.UtilityContext = utilContext
	//valMap := shared.GetPocketState().ValidatorMap
	maxTxBytes := 90000 // INTEGRATION_TEMP
	//proposer := []byte(strconv.Itoa(int(m.NodeId)))
	pk, _ := crypto.GeneratePrivateKey()
	lastByzValidators := make([][]byte, 0) // INTEGRATION_TEMP: m.UtilityContext.GetPersistanceContext().GetLastByzValidators
	txs, err := m.UtilityContext.GetTransactionsForProposal(pk.PublicKey().Address(), maxTxBytes, lastByzValidators)
	if err != nil {
		return nil, err
	}

	pocketState := types2.GetPocketState()

	header := &types.BlockHeaderConsTemp{
		Height: int64(m.Height),
		Hash:   strconv.Itoa(int(m.Height)),

		LastBlockHash:   pocketState.AppHash,
		ProposerAddress: []byte(pocketState.Address),
		// ProposerId:      uint32(m.NodeId),
		// QuorumCertificate // TODO
	}

	fmt.Println("[TODO] INTEGRATION_TEMP: Not useing txs yet: ", txs)
	block := &types.BlockConsTemp{
		BlockHeader: header,
		// Transactions:      make([]*typespb.Transaction, 0), // TODO: Use `txs` here.
		// ConsensusEvidence: make([]*typespb.Evidence, 0),
	}

	return block, nil
}

func (m *ConsensusModule) isValidBlock(block *types.BlockConsTemp) bool {
	if block == nil {
		return false
	}
	return true

}

// TODO: Should this be async?
func (m *ConsensusModule) deliverTxToUtility(block *types.BlockConsTemp) error {
	utilityModule := m.GetBus().GetUtilityModule()
	m.UtilityContext, _ = utilityModule.NewContext(int64(m.Height))
	proposer := []byte(strconv.Itoa(int(m.NodeId)))
	tx := make([][]byte, 0)                // INTEGRATION_TEMP: Get from block.Transactions
	lastByzValidators := make([][]byte, 0) // INTEGRATION_TEMP: m.UtilityContext.GetPersistanceContext().GetLastByzValidators

	appHash, err := m.UtilityContext.ApplyBlock(int64(m.Height), proposer, tx, lastByzValidators)
	if err != nil {
		return err
	}

	// INTEGRATION_TEMP: Make sure the BlockHeader uses the same encoding as the appHash
	if block.BlockHeader.Hash != hex.EncodeToString(appHash) {
		return fmt.Errorf("[ERROR] Why is the block header hash not what utility returned?")
	}

	return nil
}

func (m *ConsensusModule) commitBlock(block *types.BlockConsTemp) error {
	m.nodeLog(fmt.Sprintf("APPLYING BLOCK AT HEIGHT %d.", m.Height))
	if err := m.UtilityContext.GetPersistanceContext().Commit(); err != nil {
		return err
	}
	m.UtilityContext.ReleaseContext()
	m.UtilityContext = nil

	//utilityModule := m.GetBus().GetUtilityModule()
	//if err := utilityModule.EndBlock(nil); err != nil {
	//	m.paceMaker.InterruptRound()
	//	return err
	//}

	pocketState := types2.GetPocketState()
	pocketState.UpdateAppHash(block.BlockHeader.Hash)
	pocketState.UpdateBlockHeight(uint64(block.BlockHeader.Height))

	// TODO: Something with the persistence module?
	// persistenceModule := m.GetBus().GetpersistenceModule()

	return nil
}
