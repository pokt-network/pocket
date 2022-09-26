package persistence

import (
	"encoding/binary"
	"encoding/hex"
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/pokt-network/pocket/persistence/types"
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

func (p PostgresContext) TransactionExists(transactionHash string) (bool, error) {
	log.Println("TODO: TransactionExists not implemented")
	return false, nil
}

func (p PostgresContext) StoreTransaction(transactionProtoBytes []byte) error {
	log.Println("TODO: StoreTransaction not implemented")
	return nil
}

func (p PostgresContext) insertBlock(block *types.Block) error {
	ctx, tx, err := p.GetCtxAndTx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, types.InsertBlockQuery(block.Height, block.Hash, block.ProposerAddress, block.QuorumCertificate))
	return err
}

func (p PostgresContext) storeBlock(block *types.Block) error {
	blockBz, err := proto.Marshal(block)
	if err != nil {
		return err
	}

	return p.blockstore.Put(heightToBytes(p.Height), blockBz)
}

func (p PostgresContext) getBlock(proposerAddr []byte, quorumCert []byte) (*types.Block, error) {
	prevHash, err := p.GetBlockHash(p.Height - 1)
	if err != nil {
		return nil, err
	}

	// TODO: get this from the transactions that were store via `StoreTransaction`
	txs := make([][]byte, 0)

	block := &types.Block{
		Height:            uint64(p.Height),
		Hash:              string(p.stateHash),
		PrevHash:          string(prevHash),
		ProposerAddress:   proposerAddr,
		QuorumCertificate: quorumCert,
		Transactions:      txs,
	}

	return block, nil
}

// CLEANUP: Should this be moved to a shared directory?
func heightToBytes(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
	return heightBytes
}
