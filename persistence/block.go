package persistence

import (
	"encoding/binary"
	"encoding/hex"
	"log"

	"github.com/pokt-network/pocket/persistence/types"
)

// OPTIMIZE(team): get from blockstore or keep in memory
func (p PostgresContext) GetLatestBlockHeight() (latestHeight uint64, err error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return 0, err
	}

	err = txn.QueryRow(ctx, types.GetLatestBlockHeightQuery()).Scan(&latestHeight)
	return
}

// OPTIMIZE(team): get from blockstore or keep in cache/memory
func (p PostgresContext) GetBlockHash(height int64) ([]byte, error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}

	var hexHash string
	err = txn.QueryRow(ctx, types.GetBlockHashQuery(height)).Scan(&hexHash)
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(hexHash)
}

func (p PostgresContext) GetHeight() (int64, error) {
	return p.Height, nil
}

func (p PostgresContext) TransactionExists(transactionHash string) (bool, error) {
	log.Println("TODO: TransactionExists not implemented")
	return false, nil
}

func (p PostgresContext) StoreTransaction(transactionProtoBytes []byte) error {
	log.Println("TODO: StoreTransaction not implemented")
	return nil
}

func (p PostgresContext) InsertBlock(height uint64, hash string, proposerAddr []byte, quorumCert []byte) error {
	ctx, tx, err := p.DB.GetCtxAndTxn()
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

func (p PostgresContext) storeBlock(blockProtoBytes []byte) error {
	return p.DB.Blockstore.Put(heightToBytes(p.Height), blockProtoBytes)
}
