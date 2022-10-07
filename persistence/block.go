package persistence

import (
	"encoding/binary"
	"encoding/hex"
	"log"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
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

func (p *PostgresContext) StoreTransaction(transactionProtoBytes []byte) error {
	p.currentBlockTxs = append(p.currentBlockTxs, transactionProtoBytes)
	return nil
}

func (p *PostgresContext) insertBlock(block *types.Block) error {
	ctx, tx, err := p.GetCtxAndTx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, types.InsertBlockQuery(block.Height, block.Hash, block.ProposerAddress, block.QuorumCertificate))
	return err
}

func (p PostgresContext) storeBlock(block *types.Block) error {
	blockBz, err := codec.GetCodec().Marshal(block)
	if err != nil {
		return err
	}
	return p.blockStore.Put(HeightToBytes(p.Height), blockBz)
}

func (p *PostgresContext) prepareBlock(proposerAddr []byte, quorumCert []byte) (*types.Block, error) {
	var prevHash []byte
	if p.Height > 0 {
		var err error
		prevHash, err = p.GetBlockHash(p.Height - 1)
		if err != nil {
			return nil, err
		}
	} else {
		prevHash = []byte("HACK: get hash from genesis")
	}

	block := &types.Block{
		Height:            uint64(p.Height),
		Hash:              hex.EncodeToString(p.currentStateHash),
		PrevHash:          hex.EncodeToString(prevHash),
		ProposerAddress:   proposerAddr,
		QuorumCertificate: quorumCert,
		Transactions:      p.currentBlockTxs,
	}

	return block, nil
}

// CLEANUP: Should this be moved to a shared directory?
// Exposed for testing purposes
func HeightToBytes(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
	return heightBytes
}
