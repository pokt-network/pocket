package modules

// TxResult is an indexed transaction. It is the result of successfully executing a `Transaction` against the blockchain state.
type TxResult interface {
	GetTx() []byte                        // a serialized `Transaction` proto
	GetHeight() int64                     // the height at which the tx was applied
	GetIndex() int32                      // the transaction's index within the block (i.e. ordered by when the proposer received it in the mempool)
	GetResultCode() int32                 // 0 is no error, otherwise corresponds to error object code; // IMPROVE: Consider using enums for the result codes
	GetError() string                     // description of the error if the result code is non-zero; IMPROVE: Add a specific type fot he error code
	GetSignerAddr() string                // the address of the signer (i.e. sender) of the transaction
	GetRecipientAddr() string             // the address of the receiver of the transaction if applicable
	GetMessageType() string               // the message type contained in the transaction; must correspond to a proto that the node can can process (e.g. Stake, Unstake, Send, etc...) // IMPROVE: How do we document all the types?
	Hash() ([]byte, error)                // the hash of the tx bytes
	HashFromBytes([]byte) ([]byte, error) // same operation as `Hash`, but avoid re-serializing the tx
	Bytes() ([]byte, error)               // returns the serialized transaction bytes
	FromBytes([]byte) (TxResult, error)   // returns the deserialized transaction result
}
