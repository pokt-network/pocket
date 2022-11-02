// TECHDEBT(andrew): Move this out of shared and alongside the mempool.

package indexer

import (
	"encoding/hex"
	"fmt"

	"github.com/pokt-network/pocket/shared/codec"
	shared "github.com/pokt-network/pocket/shared/modules"

	"github.com/jordanorelli/lexnum"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/crypto"
)

// Interface

// `TxIndexer` interface defines methods to index and query transactions.
type TxIndexer interface {
	// `Index` analyzes, indexes and stores a single transaction result.
	// `Index` indexes by `(hash, height, sender, recipient)`
	Index(result shared.TxResult) error

	// `GetByHash` returns the transaction specified by the hash if indexed or nil otherwise
	GetByHash(hash []byte) (shared.TxResult, error)

	// `GetByHeight` returns all transactions specified by height or nil if there are no transactions at that height
	GetByHeight(height int64, descending bool) ([]shared.TxResult, error)

	// `GetBySender` returns all transactions signed by *sender*; may be ordered descending/ascending
	GetBySender(sender string, descending bool) ([]shared.TxResult, error)

	// GetByRecipient returns all transactions *sent to address*; may be ordered descending/ascending
	GetByRecipient(recipient string, descending bool) ([]shared.TxResult, error)

	// Close stops the underlying db connection
	Close() error
}

// Implementation
var _ TxIndexer = &txIndexer{}
var _ shared.TxResult = &TxRes{}

const (
	hashPrefix      = 'h'
	heightPrefix    = 'b' // b for block
	senderPrefix    = 's'
	recipientPrefix = 'r'
)

// =,- are the default parameters in the [example repository](https://github.com/jordanorelli/lexnum#example)
// INVESTIGATE: We can research to see if there are more optimal parameters
var elenEncoder = lexnum.NewEncoder('=', '-')

func (x *TxRes) Bytes() ([]byte, error) {
	return codec.GetCodec().Marshal(x)
}

func (*TxRes) FromBytes(bz []byte) (shared.TxResult, error) {
	result := new(TxRes)
	if err := codec.GetCodec().Unmarshal(bz, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (x *TxRes) Hash() ([]byte, error) {
	bz, err := x.Bytes()
	if err != nil {
		return nil, err
	}
	return x.HashFromBytes(bz)
}

func (x *TxRes) HashFromBytes(bz []byte) ([]byte, error) {
	return crypto.SHA3Hash(bz), nil
}

type txIndexer struct {
	db kvstore.KVStore
}

func NewTxIndexer(databasePath string) (TxIndexer, error) {
	if databasePath == "" {
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

func (indexer *txIndexer) Index(result shared.TxResult) error {
	bz, err := result.Bytes()
	if err != nil {
		return err
	}
	hash, err := result.HashFromBytes(bz)
	if err != nil {
		return err
	}
	hashKey, err := indexer.indexByHash(hash, bz)
	if err != nil {
		return err
	}
	if err := indexer.indexByHeightAndIndex(result.GetHeight(), result.GetIndex(), hashKey); err != nil {
		return err
	}
	if err := indexer.indexBySender(result.GetSignerAddr(), hashKey); err != nil {
		return err
	}
	if err := indexer.indexByRecipient(result.GetRecipientAddr(), hashKey); err != nil {
		return err
	}
	return nil
}

func (indexer *txIndexer) GetByHash(hash []byte) (shared.TxResult, error) {
	return indexer.get(indexer.hashKey(hash))
}

func (indexer *txIndexer) GetByHeight(height int64, descending bool) ([]shared.TxResult, error) {
	return indexer.getAll(indexer.heightKey(height), descending)
}

func (indexer *txIndexer) GetBySender(sender string, descending bool) ([]shared.TxResult, error) {
	return indexer.getAll(indexer.senderKey(sender), descending)
}

func (indexer *txIndexer) GetByRecipient(recipient string, descending bool) ([]shared.TxResult, error) {
	return indexer.getAll(indexer.recipientKey(recipient), descending)
}

func (indexer *txIndexer) Close() error {
	return indexer.db.Stop()
}

// kv helper functions

func (indexer *txIndexer) getAll(prefix []byte, descending bool) (result []shared.TxResult, err error) {
	_, hashKeys, err := indexer.db.GetAll(prefix, descending)
	if err != nil {
		return nil, err
	}
	for _, hashKey := range hashKeys {
		var txResult shared.TxResult
		txResult, err = indexer.get(hashKey)
		if err != nil {
			return
		}
		result = append(result, txResult)
	}
	return
}

func (indexer *txIndexer) get(key []byte) (shared.TxResult, error) {
	bz, err := indexer.db.Get(key)
	if err != nil {
		return nil, err
	}
	return new(TxRes).FromBytes(bz)
}

// index helper functions

func (indexer *txIndexer) indexByHash(hash, bz []byte) (hashKey []byte, err error) {
	key := indexer.hashKey(hash)
	return key, indexer.db.Set(key, bz)
}

func (indexer *txIndexer) indexByHeightAndIndex(height int64, index int32, bz []byte) error {
	return indexer.db.Set(indexer.heightAndIndexKey(height, index), bz)
}

func (indexer *txIndexer) indexBySender(sender string, bz []byte) error {
	return indexer.db.Set(indexer.senderKey(sender), bz)
}

func (indexer *txIndexer) indexByRecipient(recipient string, bz []byte) error {
	if recipient == "" {
		return nil
	}
	return indexer.db.Set(indexer.recipientKey(recipient), bz)
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
	return indexer.key(senderPrefix, address)
}

func (indexer *txIndexer) recipientKey(address string) []byte {
	return indexer.key(recipientPrefix, address)
}

func (indexer *txIndexer) key(prefix rune, postfix string) []byte {
	return []byte(fmt.Sprintf("%s/%s", string(prefix), postfix))
}
