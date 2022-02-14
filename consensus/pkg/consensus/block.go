package consensus

import (
	"encoding/hex"
	"fmt"
	"pocket/utility/shared/crypto"
	"strconv"

	"pocket/shared"
	"pocket/shared/typespb"
)

func (m *consensusModule) prepareBlock() (*typespb.BlockConsTemp, error) {
	if m.UtilityContext != nil {
		m.nodeLog("[WARN] Why is the node utility context not nil when preparing a new block?. Realising for now...")
		m.UtilityContext.ReleaseContext()
	}
	utilContext, err := m.GetPocketBusMod().GetUtilityModule().NewUtilityContextWrapper(int64(m.Height))
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

	pocketState := shared.GetPocketState()

	header := &typespb.BlockHeaderConsTemp{
		Height: int64(m.Height),
		Hash:   strconv.Itoa(int(m.Height)),

		LastBlockHash:   pocketState.AppHash,
		ProposerAddress: []byte(pocketState.Address),
		// ProposerId:      uint32(m.NodeId),
		// QuorumCertificate // TODO
	}

	fmt.Println("[TODO] INTEGRATION_TEMP: Not useing txs yet: ", txs)
	block := &typespb.BlockConsTemp{
		BlockHeader: header,
		// Transactions:      make([]*typespb.Transaction, 0), // TODO: Use `txs` here.
		// ConsensusEvidence: make([]*typespb.Evidence, 0),
	}

	return block, nil
}

func (m *consensusModule) isValidBlock(block *typespb.BlockConsTemp) bool {
	if block == nil {
		return false
	}
	return true

}

// TODO: Should this be async?
func (m *consensusModule) deliverTxToUtility(block *typespb.BlockConsTemp) error {
	utilityModule := m.GetPocketBusMod().GetUtilityModule()
	m.UtilityContext, _ = utilityModule.NewUtilityContextWrapper(int64(m.Height))
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

func (m *consensusModule) commitBlock(block *typespb.BlockConsTemp) error {
	m.nodeLog(fmt.Sprintf("APPLYING BLOCK AT HEIGHT %d.", m.Height))
	if err := m.UtilityContext.GetPersistanceContext().Commit(); err != nil {
		return err
	}
	m.UtilityContext.ReleaseContext()
	m.UtilityContext = nil

	//utilityModule := m.GetPocketBusMod().GetUtilityModule()
	//if err := utilityModule.EndBlock(nil); err != nil {
	//	m.paceMaker.InterruptRound()
	//	return err
	//}

	pocketState := shared.GetPocketState()
	pocketState.UpdateAppHash(block.BlockHeader.Hash)
	pocketState.UpdateBlockHeight(uint64(block.BlockHeader.Height))

	// TODO: Something with the persistence module?
	// persistenceModule := m.GetPocketBusMod().GetpersistenceModule()

	return nil
}
