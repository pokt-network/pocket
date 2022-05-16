package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/schema"
)

// TODO(team): get from blockstore or keep in memory
func (p PostgresContext) GetLatestBlockHeight() (latestHeight uint64, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	err = conn.QueryRow(ctx, schema.LatestBlockHeightQuery()).Scan(&latestHeight)
	return
}

func (p PostgresContext) GetBlockHash(height int64) ([]byte, error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}
	var hexHash string
	err = conn.QueryRow(ctx, schema.BlockHashQuery(height)).Scan(&hexHash)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(hexHash)
}

func (p PostgresContext) NewSavePoint(bytes []byte) error {
	// TODO(team): Persistence.NewSavePoint not implemented
	return nil
}

func (p PostgresContext) RollbackToSavePoint(bytes []byte) error {
	// TODO(team): Persistence.RollbackToSavePoint not implemented
	return nil
}

func (p PostgresContext) AppHash() ([]byte, error) {
	// TODO(team): Persistence.AppHash not implemented
	return []byte("this_is_a_placeholder"), nil
}

func (p PostgresContext) Reset() error {
	// TODO(team): Persistence.Reset not implemented
	return nil
}

func (p PostgresContext) Commit() error {
	//TODO(team): Persistence.Commit not implemented
	return nil
}

func (p PostgresContext) Release() {
	// TODO(team): Persistence.Release not implemented
}

func (p PostgresContext) GetHeight() (int64, error) {
	return p.Height, nil
}

func (p PostgresContext) TransactionExists(transactionHash string) bool {
	// TODO(team): Persistence.TransactionExists not implemented
	return true
}
