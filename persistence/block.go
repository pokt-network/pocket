package persistence

import (
	"encoding/hex"
	"github.com/pokt-network/pocket/persistence/schema"
)

func (p PostgresContext) GetLatestBlockHeight() (uint64, error) { // TODO get from blockstore or keep in memory
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	row, err := conn.Query(ctx, schema.LatestBlockHeightQuery())
	if err != nil {
		return 0, err
	}
	var latestHeight uint64
	err = row.Scan(&latestHeight)
	return latestHeight, err
}

func (p PostgresContext) GetBlockHash(height int64) ([]byte, error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}
	row, err := conn.Query(ctx, schema.BlockHashQuery(height))
	if err != nil {
		return nil, err
	}
	var hexHash string
	if err = row.Scan(&hexHash); err != nil {
		return nil, err
	}
	return hex.DecodeString(hexHash)
}

func (p PostgresContext) NewSavePoint(bytes []byte) error {
	//TODO implement me
	return nil
}

func (p PostgresContext) RollbackToSavePoint(bytes []byte) error {
	//TODO implement me
	return nil
}

func (p PostgresContext) AppHash() ([]byte, error) {
	//TODO implement me
	return []byte("this_is_a_placeholder"), nil
}

func (p PostgresContext) Reset() error {
	return nil
	//TODO implement me
}

func (p PostgresContext) Commit() error {
	return nil
	//TODO implement me
}

func (p PostgresContext) Release() {
	//TODO implement me
}

func (p PostgresContext) GetHeight() (int64, error) {
	return p.Height, nil
}

func (p PostgresContext) TransactionExists(transactionHash string) bool {
	return true
	//TODO implement me
}
