package persistence

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/pokt-network/pocket/persistence/trees"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (p *persistenceModule) TransactionExists(txHash, txProtoBz []byte) (bool, error) {
	exists, err := p.GetBus().GetTreeStore().Prove(trees.TransactionsTreeName, txHash, txProtoBz)
	if err != nil {
		return false, err
	}
	// exclusion proof verification
	if bytes.Equal(txProtoBz, nil) && exists {
		return false, nil
	}
	// inclusion proof verification
	return exists, nil
}

func (p *PostgresContext) GetMinimumBlockHeight() (latestHeight uint64, err error) {
	ctx, tx := p.getCtxAndTx()

	err = tx.QueryRow(ctx, types.GetMinimumBlockHeightQuery()).Scan(&latestHeight)
	return
}

func (p *PostgresContext) GetMaximumBlockHeight() (latestHeight uint64, err error) {
	ctx, tx := p.getCtxAndTx()

	err = tx.QueryRow(ctx, types.GetMaximumBlockHeightQuery()).Scan(&latestHeight)
	return
}

func (p *PostgresContext) GetBlockHash(height int64) (string, error) {
	ctx, tx := p.getCtxAndTx()

	var blockHash string
	if err := tx.QueryRow(ctx, types.GetBlockHashQuery(height)).Scan(&blockHash); err != nil {
		return "", err
	}

	return blockHash, nil
}

// TODO: Consider removing this function and using `Height` directly
func (p *PostgresContext) GetHeight() (int64, error) {
	return p.Height, nil
}

// Creates a block protobuf object using the schema defined in the persistence module
func (p *PostgresContext) prepareBlock(proposerAddr, quorumCert []byte) (*coreTypes.Block, error) {
	// Retrieve the previous block hash
	var prevBlockHash string
	if p.Height != 0 {
		var err error
		prevBlockHash, err = p.GetBlockHash(p.Height - 1)
		if err != nil {
			return nil, err
		}
	}

	// Retrieve the indexed transactions at the current height
	idxTxs, err := p.txIndexer.GetByHeight(p.Height, false)
	if err != nil {
		return nil, err
	}

	// Retrieve the transactions from the idxTxs
	txs := make([][]byte, len(idxTxs))
	for i, idxTx := range idxTxs {
		txs[i] = idxTx.GetTx()
	}

	// Get the current timestamp
	// TECHDEBT: This will lead to different timestamp in each node's block store because `prepareBlock` is called locally. Needs to be revisisted and decided on a proper implementation.
	timestamp := timestamppb.Now()

	// Get the current validator set and next validator set hashes
	currSetHash, nextSetHash, err := p.getCurrentAndNextValSetHashes()
	if err != nil {
		return nil, err
	}

	// Preapre the block proto
	blockHeader := &coreTypes.BlockHeader{
		Height:            uint64(p.Height),
		NetworkId:         p.networkId,
		StateHash:         p.stateHash,
		PrevStateHash:     prevBlockHash,
		ProposerAddress:   proposerAddr,
		QuorumCertificate: quorumCert,
		Timestamp:         timestamp,
		StateTreeHashes:   p.stateTrees.GetTreeHashes(),
		ValSetHash:        currSetHash,
		NextValSetHash:    nextSetHash,
	}
	block := &coreTypes.Block{
		BlockHeader:  blockHeader,
		Transactions: txs,
	}

	p.logger.Info().Uint64("height", block.BlockHeader.Height).Msg("Storing block in block store.")

	return block, nil
}

// Inserts the block into the postgres database
func (p *PostgresContext) insertBlock(block *coreTypes.Block) error {
	if block.BlockHeader == nil {
		return fmt.Errorf("block header is nil")
	}
	blockHeader := block.BlockHeader

	ctx, tx := p.getCtxAndTx()

	_, err := tx.Exec(ctx, types.InsertBlockQuery(blockHeader.Height, blockHeader.StateHash, blockHeader.ProposerAddress, blockHeader.QuorumCertificate))
	return err
}

// getValidatorSetHashes returns the present (current p.Height-1) and next (p.Height) validator set hashes
func (p *PostgresContext) getCurrentAndNextValSetHashes() (currentValSetHash, nextValSetHash string, err error) {
	// Get the next validator set
	nextValSetHash, err = p.hashValidatorSet(p.Height)
	if err != nil {
		return "", "", err
	}

	if p.Height == 0 {
		return "", nextValSetHash, nil
	}

	// Get the current validator set
	currentValSetHash, err = p.hashValidatorSet(p.Height - 1)
	if err != nil {
		return "", "", err
	}

	return currentValSetHash, nextValSetHash, nil
}

// hashValidatorSet hashes the validator set at the given height
func (p *PostgresContext) hashValidatorSet(height int64) (string, error) {
	valSet, err := p.GetValidatorSet(height)
	if err != nil {
		return "", err
	}
	valSetBz, err := codec.GetCodec().Marshal(valSet)
	if err != nil {
		return "", err
	}
	valSetHash := crypto.SHA3Hash(valSetBz)
	return hex.EncodeToString(valSetHash), nil
}
