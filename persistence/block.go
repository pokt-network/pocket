package persistence

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (p *persistenceModule) TransactionExists(transactionHash string) (bool, error) {
	hash, err := hex.DecodeString(transactionHash)
	if err != nil {
		return false, err
	}
	res, err := p.txIndexer.GetByHash(hash)
	if res == nil {
		// check for not found
		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
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

	// Preapre the block proto
	blockHeader := &coreTypes.BlockHeader{
		Height:            uint64(p.Height),
		NetworkId:         p.networkId,
		StateHash:         p.stateHash,
		PrevStateHash:     prevBlockHash,
		ProposerAddress:   proposerAddr,
		QuorumCertificate: quorumCert,
		Timestamp:         timestamp,
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
