package persistence

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
)

// OPTIMIZE(team): get from blockstore or keep in memory
func (p PostgresContext) GetLatestBlockHeight() (latestHeight uint64, err error) {
	ctx, tx, err := p.GetCtxAndTx()
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(ctx, types.GetLatestBlockHeightQuery()).Scan(&latestHeight)
	return
}

// OPTIMIZE(team): get from blockstore or keep in cache/memory
func (p PostgresContext) GetBlockHash(height int64) ([]byte, error) {
	ctx, tx, err := p.GetCtxAndTx()
	if err != nil {
		return nil, err
	}

	var hexHash string
	err = tx.QueryRow(ctx, types.GetBlockHashQuery(height)).Scan(&hexHash)
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(hexHash)
}

func (p PostgresContext) GetHeight() (int64, error) {
	return p.Height, nil
}

func (p PostgresContext) GetPrevAppHash() (string, error) {
	height, err := p.GetHeight()
	if err != nil {
		return "", err
	}
	if height <= 1 {
		return "TODO: get from genesis", nil
	}
	block, err := p.blockstore.Get(heightToBytes(height - 1))
	if err != nil {
		return "", fmt.Errorf("error getting block hash for height %d even though it's in the database: %s", height, err)
	}
	return hex.EncodeToString(block), nil // TODO(#284): Return `block.Hash` instead of the hex encoded representation of the blockBz
}

func (p PostgresContext) GetTxResults() []modules.TxResult {
	return p.txResults
}

func (p PostgresContext) TransactionExists(transactionHash string) (bool, error) {
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

func (p PostgresContext) indexTransactions() error {
	// TODO: store in batch
	for _, txResult := range p.GetLatestTxResults() {
		if err := p.txIndexer.Index(txResult); err != nil {
			return err
		}
	}
	return nil
}

// DISCUSS: this might be retrieved from the block store - temporarily we will access it directly from the module
//       following the pattern of the Consensus Module prior to pocket/issue-#315
// TODO(#284): Remove blockProtoBytes from the interface
func (p *PostgresContext) SetProposalBlock(blockHash string, blockProtoBytes, proposerAddr []byte, transactions [][]byte) error {
	p.blockHash = blockHash
	p.blockProtoBytes = blockProtoBytes
	p.proposerAddr = proposerAddr
	p.blockTxs = transactions
	return nil
}

// TEMPORARY: Including two functions for the SQL and KV Store as an interim solution
//                 until we include the schema as part of the SQL Store because persistence
//                 currently has no access to the protobuf schema which is the source of truth.
// TODO: atomic operations needed here - inherited pattern from consensus module
func (p PostgresContext) storeBlock(quorumCert []byte) error {
	if p.blockProtoBytes == nil {
		// IMPROVE/CLEANUP: HACK - currently tests call Commit() on the same height and it throws a
		// ERROR: duplicate key value violates unique constraint "block_pkey", because it attempts to
		// store a block at height 0 for each test. We need a cleanup function to clear the block table
		// each iteration
		return nil
	}
	// INVESTIGATE: Note that we are writing this directly to the blockStore. Depending on how
	// the use of the PostgresContext evolves, we may need to write this to `ContextStore` and copy
	// over to `BlockStore` when the block is committed.
	if err := p.blockstore.Put(heightToBytes(p.Height), p.blockProtoBytes); err != nil {
		return err
	}
	// Store in SQL Store
	if err := p.InsertBlock(uint64(p.Height), p.blockHash, p.proposerAddr, quorumCert); err != nil {
		return err
	}
	// Store transactions in indexer
	return p.indexTransactions()
}

func (p PostgresContext) InsertBlock(height uint64, hash string, proposerAddr []byte, quorumCert []byte) error {
	ctx, tx, err := p.GetCtxAndTx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, types.InsertBlockQuery(height, hash, proposerAddr, quorumCert))
	return err
}

// CLEANUP: Should this be moved to a shared directory?
func heightToBytes(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
	return heightBytes
}
