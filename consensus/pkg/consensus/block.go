package consensus

import (
	"fmt"
	"strconv"

	"pocket/consensus/pkg/shared"
	"pocket/consensus/pkg/types/typespb"
)

func (m *consensusModule) prepareBlock() (*typespb.Block, error) {
	txs, err := m.GetPocketBusMod().GetUtilityModule().ReapMempool(nil)
	if err != nil {
		return nil, err
	}

	pocketState := shared.GetPocketState()

	header := &typespb.BlockHeader{
		Height: int64(m.Height),
		Hash:   strconv.Itoa(int(m.Height)),

		LastBlockHash:   pocketState.AppHash,
		ProposerAddress: []byte(pocketState.Address),
		ProposerId:      uint32(m.NodeId),
		// QuorumCertificate // TODO
	}

	block := &typespb.Block{
		BlockHeader:       header,
		Transactions:      txs,
		ConsensusEvidence: make([]*typespb.Evidence, 0),
	}

	return block, nil
}

func (m *consensusModule) isValidBlock(block *typespb.Block) bool {
	if block == nil {
		return false
	}
	return true

}

// TODO: Should this be async?
func (m *consensusModule) deliverTxToUtility(block *typespb.Block) error {
	utilityModule := m.GetPocketBusMod().GetUtilityModule()
	if err := utilityModule.BeginBlock(nil); err != nil {
		return err
	}
	for _, tx := range block.Transactions {
		if err := utilityModule.DeliverTx(nil, tx); err != nil {
			return err
		}
	}
	return nil
}

func (m *consensusModule) commitBlock(block *typespb.Block) error {
	m.nodeLog(fmt.Sprintf("APPLYING BLOCK AT HEIGHT %d.", m.Height))

	utilityModule := m.GetPocketBusMod().GetUtilityModule()
	if err := utilityModule.EndBlock(nil); err != nil {
		m.paceMaker.InterruptRound()
		return err
	}

	pocketState := shared.GetPocketState()
	pocketState.UpdateAppHash(block.BlockHeader.Hash)
	pocketState.UpdateBlockHeight(uint64(block.BlockHeader.Height))

	// TODO: Something with the persistance module?
	// persistanceModule := m.GetPocketBusMod().GetPersistanceModule()

	return nil
}
