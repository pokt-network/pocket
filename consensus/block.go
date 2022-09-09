package consensus

import (
	"encoding/hex"
	"github.com/pokt-network/pocket/shared/codec"
	"unsafe"

	typesCons "github.com/pokt-network/pocket/consensus/types"
)

// TODO(olshansky): Sync with Andrew on the type of validation we need here.
func (m *ConsensusModule) validateBlock(block *typesCons.Block) error {
	if block == nil {
		return typesCons.ErrNilBlock
	}
	return nil
}

// This is a helper function intended to be called by a leader/validator during a view change
func (m *ConsensusModule) prepareBlockAsLeader() (*typesCons.Block, error) {
	if m.isReplica() {
		return nil, typesCons.ErrReplicaPrepareBlock
	}

	if err := m.refreshUtilityContext(); err != nil {
		return nil, err
	}

	txs, err := m.utilityContext.GetProposalTransactions(m.privateKey.Address(), maxTxBytes, lastByzValidators)
	if err != nil {
		return nil, err
	}

	appHash, err := m.utilityContext.ApplyBlock(int64(m.Height), m.privateKey.Address(), txs, lastByzValidators)
	if err != nil {
		return nil, err
	}

	blockHeader := &typesCons.BlockHeader{
		Height:            int64(m.Height),
		Hash:              hex.EncodeToString(appHash),
		NumTxs:            uint32(len(txs)),
		LastBlockHash:     m.appHash,
		ProposerAddress:   m.privateKey.Address().Bytes(),
		QuorumCertificate: []byte("HACK: Temporary placeholder"),
	}

	block := &typesCons.Block{
		BlockHeader:  blockHeader,
		Transactions: txs,
	}

	return block, nil
}

// This is a helper function intended to be called by a replica/voter during a view change
func (m *ConsensusModule) applyBlockAsReplica(block *typesCons.Block) error {
	if m.isLeader() {
		return typesCons.ErrLeaderApplyBLock
	}

	// TODO(olshansky): Add unit tests to verify this.
	if unsafe.Sizeof(*block) > uintptr(m.MaxBlockBytes) {
		return typesCons.ErrInvalidBlockSize(uint64(unsafe.Sizeof(*block)), m.MaxBlockBytes)
	}

	if err := m.refreshUtilityContext(); err != nil {
		return err
	}

	appHash, err := m.utilityContext.ApplyBlock(int64(m.Height), block.BlockHeader.ProposerAddress, block.Transactions, lastByzValidators)
	if err != nil {
		return err
	}

	// DISCUSS(drewsky): Is `ApplyBlock` going to return blockHash or appHash?
	if block.BlockHeader.Hash != hex.EncodeToString(appHash) {
		return typesCons.ErrInvalidAppHash(block.BlockHeader.Hash, hex.EncodeToString(appHash))
	}

	return nil
}

// Creates a new Utility context and clears/nullifies any previous contexts if they exist
func (m *ConsensusModule) refreshUtilityContext() error {
	// This is a catch-all to release the previous utility context if it wasn't cleaned up
	// in the proper lifecycle (e.g. catch up, error, network partition, etc...). Ideally, this
	// should not be called.
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

func (m *ConsensusModule) commitBlock(block *typesCons.Block) error {
	m.nodeLog(typesCons.CommittingBlock(m.Height, len(block.Transactions)))

	// Store the block in the KV store
	codec := codec.GetCodec()
	blockProtoBytes, err := codec.Marshal(block)
	if err != nil {
		return err
	}

	// IMPROVE(olshansky): temporary solution. `ApplyBlock` above applies the
	// transactions to the postgres database, and this stores it in the KV store upon commitment.
	// Instead of calling this directly, an alternative solution is to store the block metadata in
	// the persistence context and have `CommitPersistenceContext` do this under the hood. However,
	// additional `Block` metadata will need to be passed through and may change when we merkle the
	// state hash.
	if err := m.utilityContext.StoreBlock(blockProtoBytes); err != nil {
		return err
	}

	// Commit and release the context
	if err := m.utilityContext.CommitPersistenceContext(); err != nil {
		return err
	}

	m.utilityContext.ReleaseContext()
	m.utilityContext = nil

	m.appHash = block.BlockHeader.Hash

	return nil
}
