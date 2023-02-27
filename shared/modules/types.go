package modules

// TxResult is a hydrated/blown-up Transaction.
// It is the result of a transaction on which basic validation has been applied, and from which the
// embedded Message, and its contents, were deserialized and extracted.
type TxResult interface {
	GetTx() []byte                        // a serialized `Transaction` proto
	GetHeight() int64                     // the block height at which the transaction was included
	GetIndex() int32                      // the transaction's index within the block (i.e. ordered by when the proposer received it in the mempool)
	GetResultCode() int32                 // 0 is no error, otherwise corresponds to error object code; // IMPROVE: Consider using enums for the result codes
	GetError() string                     // description of the error if the result code is non-zero; IMPROVE: Add a specific type fot he error code
	GetSignerAddr() string                // the address of the signer (e.g. sender) of the transaction
	GetRecipientAddr() string             // Optional: the address of the recipient of the transaction (if applicable)
	GetMessageType() string               // the message type contained in the transaction; must correspond to a proto that the node can can process (e.g. Stake, Unstake, Send, etc...) // IMPROVE: How do we document all the types?
	Bytes() ([]byte, error)               // returns the serialized `TxResult`
	FromBytes([]byte) (TxResult, error)   // returns the deserialized `TxResult`
	Hash() ([]byte, error)                // the hash of the TxResult bytes
	HashFromBytes([]byte) ([]byte, error) // same operation as `Hash`, but avoid re-serializing the tx
}
