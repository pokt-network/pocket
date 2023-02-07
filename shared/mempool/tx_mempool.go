package mempool

type TXMempool interface {
	Contains(hash string) bool
	AddTx(tx []byte) error
	RemoveTx(tx []byte) error

	Clear()
	TxCount() uint32 // Returns the number of transactions stored in the mempool
	IsEmpty() bool
	TxsBytesTotal() uint64 // Returns the total sum of all transactions' sizes (in bytes) stored in the mempool
	PopTx() (tx []byte, err error)
}
