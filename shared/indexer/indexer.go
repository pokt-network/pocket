// TECHDEBT(andrew): Move this out of shared and alongside the mempool.

package indexer

import (
	"encoding/hex"
	"fmt"

	"github.com/pokt-network/pocket/shared/codec"

	"github.com/jordanorelli/lexnum"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/crypto"
)

// Interface

// `TxIndexer` interface defines methods to index and query transactions.
// TODO: Link to the `bus` module
type TxIndexer interface {
	// `Index` analyzes, indexes and stores a single transaction result.
	// `Index` indexes by `(hash, height, sender, recipient)`
	Index(result TxResult) error

	// `GetByHash` returns the transaction specified by the hash if indexed or nil otherwise
	GetByHash(hash []byte) (TxResult, error)

	// `GetByHeight` returns all transactions specified by height or nil if there are no transactions at that height
	GetByHeight(height int64, descending bool) ([]TxResult, error)

	// `GetBySender` returns all transactions signed by *sender*; may be ordered descending/ascending
	GetBySender(sender string, descending bool) ([]TxResult, error)

	// GetByRecipient returns all transactions *sent to address*; may be ordered descending/ascending
	GetByRecipient(recipient string, descending bool) ([]TxResult, error)

	// Close stops the underlying db connection
	Close() error
}

// The result of executing a transaction against the blockchain state so that it is included in the block
type TxResult interface {
	GetTx() []byte                        // the transaction object primitive
	GetHeight() int64                     // the height at which the tx was applied
	GetIndex() int32                      // the transaction's index within the block (i.e. ordered by when the proposer received it in the mempool)
	GetResultCode() int32                 // 0 is no error, otherwise corresponds to error object code; // IMPROVE: Add a specific type fot he result code
	GetError() string                     // can be empty; IMPROVE: Add a specific type fot he error code
	GetSignerAddr() string                // get the address of who signed (i.e. sent) the transaction
	GetRecipientAddr() string             // get the address of who received the transaction; may be empty
	GetMessageType() string               // corresponds to type of message (validator-stake, app-unjail, node-stake, etc) // IMPROVE: Add an enum for message types
	Hash() ([]byte, error)                // the hash of the tx bytes
	HashFromBytes([]byte) ([]byte, error) // same operation as `Hash`, but avoid re-serializing the tx
	Bytes() ([]byte, error)               // returns the serialized transaction bytes
	FromBytes([]byte) (TxResult, error)   // returns the deserialized transaction result
}

// Implementation

var _ TxResult = &DefaultTxResult{}
var _ TxIndexer = &txIndexer{}

// TODO(andrew): Move this documentation to a README.

// `txIndexer` implementation uses a `KVStore` (interface) to index the transactions
//
// The transaction is indexed in the following formats:
// - HASHKEY:      "h/SHA3(TxResultProtoBytes)"  VAL: TxResultProtoBytes     // store value by hash (the key here is equivalent to the VALs below)
// - HEIGHTKEY:    "b/height/index"              VAL: HASHKEY                // store hashKey by height
// - SENDERKEY:    "s/senderAddr"                VAL: HASHKEY                // store hashKey by sender
// - RECIPIENTKEY: "r/recipientAddr"             VAL: HASHKEY                // store hashKey by recipient (if not empty)
//
// FOOTNOTE: the height/index store is using [ELEN](https://github.com/jordanorelli/lexnum/blob/master/elen.pdf)
// This is to ensure the results are stored sorted (assuming the `KVStore`` uses a byte-wise lexicographical sorting)

const (
	hashPrefix      = 'h'
	heightPrefix    = 'b' // b for block
	senderPrefix    = 's'
	recipientPrefix = 'r'
)

// =,- are the default parameters in the [example repository](https://github.com/jordanorelli/lexnum#example)
// INVESTIGATE: We can research to see if there are more optimal parameters
var elenEncoder = lexnum.NewEncoder('=', '-')

func (x *DefaultTxResult) Bytes() ([]byte, error) {
	return codec.GetCodec().Marshal(x)
}

func (*DefaultTxResult) FromBytes(bz []byte) (TxResult, error) {
	result := new(DefaultTxResult)
	if err := codec.GetCodec().Unmarshal(bz, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (x *DefaultTxResult) Hash() ([]byte, error) {
	bz, err := x.Bytes()
	if err != nil {
		return nil, err
	}
	return x.HashFromBytes(bz)
}

func (x *DefaultTxResult) HashFromBytes(bz []byte) ([]byte, error) {
	return crypto.SHA3Hash(bz), nil
}

type txIndexer struct {
	db kvstore.KVStore
}

func NewTxIndexer(databasePath string) (TxIndexer, error) {
	db, err := kvstore.OpenKVStore(databasePath)
	return &txIndexer{
		db: db,
	}, err
}

func NewMemTxIndexer() (TxIndexer, error) {
	return &txIndexer{
		db: kvstore.NewMemKVStore(),
	}, nil
}

func (indexer *txIndexer) Index(result TxResult) error {
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

func (indexer *txIndexer) GetByHash(hash []byte) (TxResult, error) {
	return indexer.get(indexer.hashKey(hash))
}

func (indexer *txIndexer) GetByHeight(height int64, descending bool) ([]TxResult, error) {
	return indexer.getAll(indexer.heightKey(height), descending)
}

func (indexer *txIndexer) GetBySender(sender string, descending bool) ([]TxResult, error) {
	return indexer.getAll(indexer.senderKey(sender), descending)
}

func (indexer *txIndexer) GetByRecipient(recipient string, descending bool) ([]TxResult, error) {
	return indexer.getAll(indexer.recipientKey(recipient), descending)
}

func (indexer *txIndexer) Close() error {
	return indexer.db.Stop()
}

// kv helper functions

func (indexer *txIndexer) getAll(prefix []byte, descending bool) (result []TxResult, err error) {
	hashKeys, err := indexer.db.GetAll(prefix, descending)
	if err != nil {
		return nil, err
	}
	for _, hashKey := range hashKeys {
		var txResult TxResult
		txResult, err = indexer.get(hashKey)
		if err != nil {
			return
		}
		result = append(result, txResult)
	}
	return
}

func (indexer *txIndexer) get(key []byte) (TxResult, error) {
	bz, err := indexer.db.Get(key)
	if err != nil {
		return nil, err
	}
	return new(DefaultTxResult).FromBytes(bz)
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
