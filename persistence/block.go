package persistence

import (
	"encoding/hex"
	"log"

	"github.com/pokt-network/pocket/persistence/schema"
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

func (p PostgresContext) TransactionExists(transactionHash string) bool {
	log.Println("TODO: Block - TransactionExists not implemented")
	return true
}
