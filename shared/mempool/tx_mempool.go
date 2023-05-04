package mempool

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

type TXMempool interface {
	Contains(hash string) bool
	AddTx(tx []byte) coreTypes.Error
	RemoveTx(tx []byte) error
	GetAll() [][]byte
	Get(hash string) []byte

	Clear()
	TxCount() uint32 // Returns the number of transactions stored in the mempool
	IsEmpty() bool
	TxsBytesTotal() uint64 // Returns the total sum of all transactions' sizes (in bytes) stored in the mempool
	PopTx() (tx []byte, err error)
}
