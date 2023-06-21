package indexer

//go:generate mockgen -package=mock_types -destination=../types/mocks/indexer_mock.go github.com/pokt-network/pocket/persistence/indexer TxIndexer

import (
	"encoding/hex"
	"fmt"

	"github.com/jordanorelli/lexnum"
	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// `TxIndexer` interface defines methods to index and query transactions.
type TxIndexer interface {
	// `Index` analyzes, indexes and stores a single transaction result.
	// `Index` indexes by `(hash, height, sender, recipient)`
	Index(result *coreTypes.IndexedTransaction) error

	// `GetByHash` returns the transaction specified by the hash if indexed or nil otherwise
	GetByHash(hash []byte) (*coreTypes.IndexedTransaction, error)

	// `GetByHeight` returns all transactions specified by height or nil if there are no transactions at that height; may be ordered descending/ascending
	GetByHeight(height int64, descending bool) ([]*coreTypes.IndexedTransaction, error)

	// `GetBySender` returns all transactions signed by *sender*; may be ordered descending/ascending
	GetBySender(sender string, descending bool) ([]*coreTypes.IndexedTransaction, error)

	// GetByRecipient returns all transactions *sent to address*; may be ordered descending/ascending
	GetByRecipient(recipient string, descending bool) ([]*coreTypes.IndexedTransaction, error)

	// Transaction atomicity
	Prepare() error
	Commit() error

	// Close stops the underlying db connection
	Close() error
}

// Implementation
var _ TxIndexer = &txIndexer{}

const (
	hashPrefix      = 'h'
	heightPrefix    = 'b' // b for block
	senderPrefix    = 's'
	recipientPrefix = 'r'
)

// =,- are the default parameters in the [example repository](https://github.com/jordanorelli/lexnum#example)
// INVESTIGATE: We can research to see if there are more optimal parameters
var elenEncoder = lexnum.NewEncoder('=', '-')

type txIndexer struct {
	db kvstore.KVStore
}

func NewTxIndexer(databasePath string) (TxIndexer, error) {
	if databasePath == ":memory:" {
		return NewMemTxIndexer()
	}

	db, err := kvstore.NewKVStore(databasePath)
	return &txIndexer{
		db: db,
	}, err
}

func NewMemTxIndexer() (TxIndexer, error) {
	return &txIndexer{
		db: kvstore.NewMemKVStore(),
	}, nil
}

func (indexer *txIndexer) Index(result *coreTypes.IndexedTransaction) error {
	bz, err := result.Bytes()
	if err != nil {
		return err
	}
	hash := result.HashFromBytes(result.GetTx())
	hashKey, err := indexer.indexByHash(hash, bz)
	if err != nil {
		return err
	}
	if err := indexer.indexByHeightAndIndex(result.GetHeight(), result.GetIndex(), hashKey); err != nil {
		return err
	}
	if err := indexer.indexBySenderHeightAndBlockIndex(result.GetSignerAddr(), result.GetHeight(), result.GetIndex(), hashKey); err != nil {
		return err
	}
	if err := indexer.indexByRecipientHeightAndBlockIndex(result.GetRecipientAddr(), result.GetHeight(), result.GetIndex(), hashKey); err != nil {
		return err
	}
	return nil
}

func (indexer *txIndexer) GetByHash(hash []byte) (*coreTypes.IndexedTransaction, error) {
	return indexer.get(indexer.hashKey(hash))
}

func (indexer *txIndexer) GetByHeight(height int64, descending bool) ([]*coreTypes.IndexedTransaction, error) {
	return indexer.getAll(indexer.heightKey(height), descending)
}

func (indexer *txIndexer) GetBySender(sender string, descending bool) ([]*coreTypes.IndexedTransaction, error) {
	return indexer.getAll(indexer.senderKey(sender), descending)
}

func (indexer *txIndexer) GetByRecipient(recipient string, descending bool) ([]*coreTypes.IndexedTransaction, error) {
	return indexer.getAll(indexer.recipientKey(recipient), descending)
}

// Atmic transaction handling

func (store *txIndexer) Prepare() error {
	return fmt.Errorf("not impl")
}

func (store *txIndexer) Commit() error {
	return fmt.Errorf("not impl")
}

func (indexer *txIndexer) Close() error {
	return indexer.db.Stop()
}

// kv helper functions

func (indexer *txIndexer) getAll(prefix []byte, descending bool) (result []*coreTypes.IndexedTransaction, err error) {
	_, hashKeys, err := indexer.db.GetAll(prefix, descending)
	if err != nil {
		return nil, err
	}
	for _, hashKey := range hashKeys {
		idxTx, err := indexer.get(hashKey)
		if err != nil {
			return nil, err
		}
		result = append(result, idxTx)
	}
	return
}

func (indexer *txIndexer) get(key []byte) (*coreTypes.IndexedTransaction, error) {
	bz, err := indexer.db.Get(key)
	if err != nil {
		return nil, err
	}
	return new(coreTypes.IndexedTransaction).FromBytes(bz)
}

// index helper functions

func (indexer *txIndexer) indexByHash(hash, bz []byte) (hashKey []byte, err error) {
	key := indexer.hashKey(hash)
	return key, indexer.db.Set(key, bz)
}

func (indexer *txIndexer) indexByHeightAndIndex(height int64, index int32, bz []byte) error {
	return indexer.db.Set(indexer.heightAndIndexKey(height, index), bz)
}

func (indexer *txIndexer) indexBySenderHeightAndBlockIndex(sender string, height int64, blockIndex int32, bz []byte) error {
	if sender == "" {
		return nil
	}
	return indexer.db.Set(indexer.senderHeightAndBlockIndexKey(sender, height, blockIndex), bz)
}

func (indexer *txIndexer) indexByRecipientHeightAndBlockIndex(recipient string, height int64, blockIndex int32, bz []byte) error {
	if recipient == "" {
		return nil
	}
	return indexer.db.Set(indexer.recipientHeightAndBlockIndexKey(recipient, height, blockIndex), bz)
}

// key helper functions

func (indexer *txIndexer) hashKey(hash []byte) []byte {
	return indexer.key(hashPrefix, hex.EncodeToString(hash))
}

func (indexer *txIndexer) heightAndIndexKey(height int64, index int32) []byte {
	return append(indexer.heightKey(height), []byte(elenEncoder.EncodeInt(int(index)))...)
}

func (indexer *txIndexer) heightKey(height int64) []byte {
	return indexer.key(heightPrefix, elenEncoder.EncodeInt(int(height))+"/")
}

func (indexer *txIndexer) senderKey(address string) []byte {
	return indexer.key(senderPrefix, address+"/")
}

func (indexer *txIndexer) senderHeightAndBlockIndexKey(address string, height int64, blockIndex int32) []byte {
	key := indexer.senderKey(address)
	key = append(key, []byte(elenEncoder.EncodeInt(int(height))+"/")...)
	key = append(key, []byte(elenEncoder.EncodeInt(int(blockIndex)))...)
	return key
}

func (indexer *txIndexer) recipientKey(address string) []byte {
	return indexer.key(recipientPrefix, address+"/")
}

func (indexer *txIndexer) recipientHeightAndBlockIndexKey(address string, height int64, blockIndex int32) []byte {
	key := indexer.recipientKey(address)
	key = append(key, []byte(elenEncoder.EncodeInt(int(height))+"/")...)
	key = append(key, []byte(elenEncoder.EncodeInt(int(blockIndex)))...)
	return key
}

func (indexer *txIndexer) key(prefix rune, postfix string) []byte {
	return []byte(fmt.Sprintf("%s/%s", string(prefix), postfix))
}
