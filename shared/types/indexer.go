package types

import (
	"encoding/hex"
	"fmt"
	"github.com/jordanorelli/lexnum"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/crypto"
)

// TODO (Team) this is only going here in shared temporarily. It should go where the mepool goes (likely Utility module) in #163

// Interfaces

type TxIndexer interface { // TxIndexer interface defines methods to index and query transactions.
	// Index analyzes, indexes and stores a single transaction result.
	// Index, indexes by hash, height, sender, and recipient
	Index(result TxResult) error

	// GetByHash returns the transaction specified by hash or nil if the transaction is not indexed
	GetByHash(hash []byte) (TxResult, error)

	// GetByHeight returns the transactions specified by height or nil if there are no transactions at that height
	GetByHeight(height int64) ([]TxResult, error)

	// GetBySender returns the transactions signed by *sender* may be ordered descending/ascending
	GetBySender(sender string, descending bool) ([]TxResult, error)

	// GetByRecipient returns the transactions *sent to address* may be ordered descending/ascending
	GetByRecipient(rec string, descending bool) ([]TxResult, error)

	// Close stops the underlying db connection
	Close() error
}

type TxResult interface {
	GetTx() []byte                        // the transaction object primitive
	GetHeight() int64                     // height it was sent
	GetIndex() int32                      // which index it was within the block-transactions
	GetResultCode() int32                 // 0 is no error, otherwise corresponds to error object code
	GetError() string                     // can be empty
	GetSender() string                    // get the address who signed
	GetRecipient() string                 // can be empty
	GetMessageType() string               // corresponds to type of message (Ex. validator-stake, app-unjail, node-stake) etc.
	Hash() ([]byte, error)                // the hash of the tx bytes
	HashFromBytes([]byte) ([]byte, error) // the hash of the tx bytes, avoids re-marshalling
	Bytes() ([]byte, error)               // proto marshalled bytes
	FromBytes([]byte) (TxResult, error)   // from proto marshalled bytes
}

var _ TxResult = &DefaultTxResult{}
var _ TxIndexer = &DefaultTxIndexer{}

// Implementation

// DefaultTxIndexer implementation uses a KVStore (interface) to index the transactions
//
// The transaction is indexed in the following formats:
// - HASHKEY:      "h/SHA3(TxResultProtoBytes)"  VAL: TxResultProtoBytes     // store value by hash
// - HEIGHTKEY:    "b/height/index"              VAL: HASHKEY                // store hashKey by height
// - SENDERKEY:    "s/height/index"              VAL: HASHKEY                // store hashKey by sender
// - RECIPIENTKEY: "r/height/index"              VAL: HASHKEY                // store hashKey by recipient (if not empty)
//
// FOOTNOTE: the height/index is store using [ELEN](https://github.com/jordanorelli/lexnum/blob/master/elen.pdf)
// This is to ensure the results are stored sorted (assuming KVStore uses a byte-wise lexicographical sorting)
//
// TODO (Team) follow up tasks
//   - Pagination for GetAll()
//   - Connect to bus module

const (
	HashPrefix            = 'h'
	HeightPrefix          = 'b' // b for block
	SenderPrefix          = 's'
	RecipientPrefix       = 'r'
	DefaultHeightOrdering = true
)

var elenEncoder = lexnum.NewEncoder('=', '-')

func (x *DefaultTxResult) Bytes() ([]byte, error) {
	return GetCodec().Marshal(x)
}

func (*DefaultTxResult) FromBytes(bz []byte) (TxResult, error) {
	result := new(DefaultTxResult)
	if err := GetCodec().Unmarshal(bz, result); err != nil {
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

type DefaultTxIndexer struct {
	db kvstore.KVStore
}

func (d *DefaultTxIndexer) Close() error {
	return d.db.Stop()
}

func NewTxIndexer(databasePath string) (TxIndexer, error) {
	db, err := kvstore.NewKVStore(databasePath)
	return &DefaultTxIndexer{
		db: db,
	}, err
}

func NewMemTxIndexer() (TxIndexer, error) {
	return &DefaultTxIndexer{
		db: kvstore.NewMemKVStore(),
	}, nil
}

func (d *DefaultTxIndexer) Index(result TxResult) error {
	bz, err := result.Bytes()
	if err != nil {
		return err
	}
	hash, err := result.HashFromBytes(bz)
	if err != nil {
		return err
	}
	hashKey, err := d.indexByHash(hash, bz)
	if err != nil {
		return err
	}
	if err := d.indexByHeightAndIndex(result.GetHeight(), result.GetIndex(), hashKey); err != nil {
		return err
	}
	if err := d.indexBySender(result.GetSender(), hashKey); err != nil {
		return err
	}
	if err := d.indexByRecipient(result.GetRecipient(), hashKey); err != nil {
		return err
	}
	return nil
}

func (d *DefaultTxIndexer) GetByHash(hash []byte) (TxResult, error) {
	return d.get(d.hashKey(hash))
}

func (d *DefaultTxIndexer) GetByHeight(height int64) ([]TxResult, error) {
	return d.getAll(d.heightKey(height), DefaultHeightOrdering)
}

func (d *DefaultTxIndexer) GetBySender(sender string, descending bool) ([]TxResult, error) {
	return d.getAll(d.senderKey(sender), descending)
}

func (d *DefaultTxIndexer) GetByRecipient(rec string, descending bool) ([]TxResult, error) {
	return d.getAll(d.recipientKey(rec), descending)
}

// kv helper functions

func (d *DefaultTxIndexer) getAll(prefix []byte, descending bool) (res []TxResult, err error) {
	hashKeys, err := d.db.GetAll(prefix, descending)
	if err != nil {
		return nil, err
	}
	for _, hashKey := range hashKeys {
		var txResult TxResult
		txResult, err = d.get(hashKey)
		if err != nil {
			return
		}
		res = append(res, txResult)
	}
	return
}

func (d *DefaultTxIndexer) get(key []byte) (TxResult, error) {
	bz, err := d.db.Get(key)
	if err != nil {
		return nil, err
	}
	return new(DefaultTxResult).FromBytes(bz)
}

// index helper functions

func (d *DefaultTxIndexer) indexByHash(hash, bz []byte) (hashKey []byte, err error) {
	key := d.hashKey(hash)
	return key, d.db.Put(key, bz)
}

func (d *DefaultTxIndexer) indexByHeightAndIndex(height int64, index int32, bz []byte) error {
	return d.db.Put(d.heightAndIndexKey(height, index), bz)
}

func (d *DefaultTxIndexer) indexBySender(sender string, bz []byte) error {
	return d.db.Put(d.senderKey(sender), bz)
}

func (d *DefaultTxIndexer) indexByRecipient(recipient string, bz []byte) error {
	if recipient == "" {
		return nil
	}
	return d.db.Put(d.recipientKey(recipient), bz)
}

// key helper functions

func (d *DefaultTxIndexer) hashKey(hash []byte) []byte {
	return d.key(HashPrefix, hex.EncodeToString(hash))
}

func (d *DefaultTxIndexer) heightAndIndexKey(height int64, index int32) []byte {
	return append(d.heightKey(height), []byte(elenEncoder.EncodeInt(int(index)))...)
}

func (d *DefaultTxIndexer) heightKey(height int64) []byte {
	return d.key(HeightPrefix, elenEncoder.EncodeInt(int(height))+"/")
}

func (d *DefaultTxIndexer) senderKey(address string) []byte {
	return d.key(SenderPrefix, address)
}

func (d *DefaultTxIndexer) recipientKey(address string) []byte {
	return d.key(RecipientPrefix, address)
}

func (d *DefaultTxIndexer) key(prefix rune, postfix string) []byte {
	return []byte(fmt.Sprintf("%s/%s", string(prefix), postfix))
}
