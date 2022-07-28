package persistence

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/pokt-network/pocket/persistence/schema"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// OPTIMIZE(team): get from blockstore or keep in cache/memory
func (p PostgresContext) GetLatestBlockHeight() (latestHeight uint64, err error) {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}

	err = conn.QueryRow(ctx, schema.GetLatestBlockHeightQuery()).Scan(&latestHeight)
	return
}

// OPTIMIZE(team): get from blockstore or keep in cache/memory
func (p PostgresContext) GetBlockHash(height int64) ([]byte, error) {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}

	var hexHash string
	err = conn.QueryRow(ctx, schema.GetBlockHashQuery(height)).Scan(&hexHash)
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(hexHash)
}

func (p PostgresContext) GetHeight() (int64, error) {
	return p.Height, nil
}

func (p PostgresContext) TransactionExists(transactionHash string) (bool, error) {
	return p.ContextStore.Exists([]byte(transactionHash))
}

func (p PostgresContext) StoreTransaction(transactionProtoBytes []byte) error {
	txHash := typesUtil.TransactionHash(transactionProtoBytes)
	return p.ContextStore.Put([]byte(txHash), transactionProtoBytes)
}

func (p PostgresContext) StoreBlock(blockProtoBytes []byte) error {
	fmt.Println("committing height", p.Height)
	// TODO_IN_THIS_COMMIT: Need to use the ContextStore and transfer over the data from the temp KV Store
	return p.BlockStore.Put(heightToBytes(p.Height), blockProtoBytes)
	// return p.ContextStore.Put(heightToBytes(p.Height), blockProtoBytes)
}

func (p PostgresContext) InsertBlock(height uint64, hash string, proposerAddr []byte, quorumCert []byte, transactions [][]byte) error {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, schema.InsertBlockQuery(height, hash, proposerAddr, quorumCert, transactions))
	return err
}

// CLEANUP: Should this be moved to a shared directory?
func heightToBytes(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
	return heightBytes
}
