package persistence

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
)

func (p *persistenceModule) TransactionExists(transactionHash string) (bool, error) {
	hash, err := hex.DecodeString(transactionHash)
	if err != nil {
		return false, err
	}
	res, err := p.txIndexer.GetByHash(hash)
	if res == nil {
		// check for not found
		if err != nil && err.Error() == kvstore.BadgerKeyNotFoundError {
			return false, nil
		}
		return false, err
	}
	return true, err
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

func (p *PostgresContext) GetHeight() (int64, error) {
	return p.Height, nil
}

// Creates a block protobuf object using the schema defined in the persistence module
func (p *PostgresContext) prepareBlock(proposerAddr, quorumCert []byte) (*coreTypes.Block, error) {
	var prevBlockHash string
	if p.Height != 0 {
		var err error
		prevBlockHash, err = p.GetBlockHash(p.Height - 1)
		if err != nil {
			return nil, err
		}
	}

	txsHash, err := p.getTxsHash()
	if err != nil {
		return nil, err
	}

	blockHeader := &coreTypes.BlockHeader{
		Height:            uint64(p.Height),
		StateHash:         p.stateHash,
		PrevStateHash:     prevBlockHash,
		ProposerAddress:   proposerAddr,
		QuorumCertificate: quorumCert,
		TransactionsHash:  txsHash,
	}
	block := &coreTypes.Block{
		BlockHeader: blockHeader,
	}

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

// Stores the block in the key-value store
func (p *PostgresContext) storeBlock(block *coreTypes.Block) error {
	blockBz, err := codec.GetCodec().Marshal(block)
	if err != nil {
		return err
	}
	log.Printf("Storing block %d in block store.\n", block.BlockHeader.Height)
	return p.blockStore.Set(utils.HeightToBytes(uint64(p.Height)), blockBz)
}
