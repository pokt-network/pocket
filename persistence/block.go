package persistence

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
)

// OPTIMIZE(team): get from blockstore or keep in memory
func (p PostgresContext) GetLatestBlockHeight() (latestHeight uint64, err error) {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(ctx, types.GetLatestBlockHeightQuery()).Scan(&latestHeight)
	return
}

// OPTIMIZE: get from blockstore or keep in cache/memory
func (p PostgresContext) GetBlockHash(height int64) ([]byte, error) {
	ctx, tx, err := p.getCtxAndTx()
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

// DISCUSS: this might be retrieved from the block store - temporarily we will access it directly from the module
//       following the pattern of the Consensus Module prior to pocket/issue-#315
// TODO(#284): Remove blockProtoBytes from the interface
func (p *PostgresContext) SetProposalBlock(blockHash string, proposerAddr, quorumCert []byte, transactions [][]byte) error {
	p.blockHash = blockHash
	p.quorumCert = quorumCert
	p.proposerAddr = proposerAddr
	p.blockTxs = transactions
	return nil
}

// Creates a block protobuf object using the schema defined in the persistence module
func (p *PostgresContext) prepareBlock(quorumCert []byte) (*types.Block, error) {
	var prevHash []byte
	if p.Height == 0 {
		prevHash = []byte("")
	} else {
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
		StateHash:         p.blockHash,
		PrevStateHash:     hex.EncodeToString(prevHash),
		ProposerAddress:   p.proposerAddr,
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
