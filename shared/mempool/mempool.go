package mempool

type TXMempool interface {
	Contains(hash string) bool
	AddTransaction(tx []byte) error
	RemoveTransaction(tx []byte) error

	Clear()
	Size() uint32 // Returns the number of transactions stored in the mempool
	IsEmpty() bool
	TxsBytes() uint64 // Returns the total sum of all transactions' sizes (in bytes) stored in the mempool
	PopTransaction() (tx []byte, err error)
}
