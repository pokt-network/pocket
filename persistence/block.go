package persistence

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/pokt-network/pocket/persistence/schema"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// OPTIMIZE(team): get from blockstore or keep in cache/memory
func (p PostgresContext) GetLatestBlockHeight() (latestHeight int64, err error) {
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
	// INVESTIGATE: Note that we are writing this directly to the blockStore. Depending on how
	// the use of the PostgresContext evolves, we may need to write this to `ContextStore` and copy
	// over to `BlockStore` when the block is committed.
	return p.BlockStore.Put(heightToBytes(p.Height), blockProtoBytes)
}

func (p PostgresContext) InsertBlock(height uint64, hash string, proposerAddr []byte, quorumCert []byte) error {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return err
	}
	fmt.Println("OLSH", hash, proposerAddr, quorumCert)
	// hash := crypto.SHA3Hash(txBz)
	// hashString := hex.EncodeToString(hash)

	// _, err = conn.Exec(ctx, schema.InsertBlockQuery(height, []byte(hash), proposerAddr, quorumCert))
	_, err = conn.Exec(ctx, schema.InsertBlockQuery(height, hex.EncodeToString([]byte("a")), []byte("b"), []byte("c")))
	return err
}

// CLEANUP: Should this be moved to a shared directory?
func heightToBytes(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
	return heightBytes
}
