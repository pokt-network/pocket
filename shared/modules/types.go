package modules

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
