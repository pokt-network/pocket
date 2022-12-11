package persistence

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
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

func (p PostgresContext) GetLatestBlockHeight() (latestHeight uint64, err error) {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(ctx, types.GetLatestBlockHeightQuery()).Scan(&latestHeight)
	return
}

func (p PostgresContext) GetBlockHash(height int64) (string, error) {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return "", err
	}

	var stateHash string
	if err = tx.QueryRow(ctx, types.GetBlockHashQuery(height)).Scan(&stateHash); err != nil {
		return "", err
	}

	return stateHash, nil
}

func (p PostgresContext) GetHeight() (int64, error) {
	return p.Height, nil
}

// Creates a block protobuf object using the schema defined in the persistence module
func (p *PostgresContext) prepareBlock(proposerAddr, quorumCert []byte) (*types.Block, error) {
	var prevHash string
	if p.Height != 0 {
		var err error
		prevHash, err = p.GetBlockHash(p.Height - 1)
		if err != nil {
			return nil, err
		}
	}

	txsHash, err := p.getTxsHash()
	if err != nil {
		return nil, err
	}

	block := &types.Block{
		Height:            uint64(p.Height),
		StateHash:         string(p.stateHash),
		PrevStateHash:     prevHash,
		ProposerAddress:   proposerAddr,
		QuorumCertificate: quorumCert,
		TransactionsHash:  txsHash,
	}

	return block, nil
}

// Inserts the block into the postgres database
func (p *PostgresContext) insertBlock(block *types.Block) error {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, types.InsertBlockQuery(block.Height, block.StateHash, block.ProposerAddress, block.QuorumCertificate))
	return err
}

// Stores the block in the key-value store
func (p PostgresContext) storeBlock(block *types.Block) error {
	blockBz, err := codec.GetCodec().Marshal(block)
	if err != nil {
		return err
	}
	return p.blockStore.Set(heightToBytes(p.Height), blockBz)
}

func heightToBytes(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
	return heightBytes
}
