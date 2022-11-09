package persistence

// TECHDEBT: Figure out why the receivers here aren't pointers?

import (
	"context"
	"log"

	"github.com/pokt-network/pocket/shared/modules"
)

func (p PostgresContext) NewSavePoint(bytes []byte) error {
	log.Println("TODO: NewSavePoint not implemented")
	return nil
}

// TODO(#327): When implementing save points and rollbacks, make sure that `prepareBlock`,
// `insertBlock`, and `storeBlock` are all atomic.
func (p PostgresContext) RollbackToSavePoint(bytes []byte) error {
	log.Println("TODO: RollbackToSavePoint not fully implemented")
	return p.GetTx().Rollback(context.TODO())
}

func (p *PostgresContext) ComputeAppHash() ([]byte, error) {
	// DISCUSS_IN_THIS_COMMIT:
	// 1. Should we compare the `appHash` returned from `updateMerkleTrees`?
	// 2. Should this update the internal state of the context?
	// Proposal: If the current internal appHash is not set, we update it.
	//           If the current internal appHash is set, we compare it and return an error if different.
	return p.updateMerkleTrees()
}

func (p PostgresContext) Commit(quorumCert []byte) error {
	log.Printf("About to commit context at height %d.\n", p.Height)

	// Create a persistence block proto
	block, err := p.prepareBlock(quorumCert)
	if err != nil {
		return err
	}

	// Store block in the KV store
	if err := p.storeBlock(block); err != nil {
		return err
	}

	// Insert the block into the SQL DB
	if err := p.insertBlock(block); err != nil {
		return err
	}

	// Commit the SQL transaction
	ctx := context.TODO()
	if err := p.GetTx().Commit(ctx); err != nil {
		return err
	}
	if err := p.conn.Close(ctx); err != nil {
		log.Println("[TODO][ERROR] Implement connection pooling. Error when closing DB connecting...", err)
	}

	return nil
}

func (p PostgresContext) Release() error {
	log.Printf("About to release context at height %d.\n", p.Height)
	ctx := context.TODO()
	if err := p.GetTx().Rollback(ctx); err != nil {
		return err
	}
	if err := p.resetContext(); err != nil {
		return err
	}
	return nil
}

func (p PostgresContext) Close() error {
	log.Printf("About to close context at height %d.\n", p.Height)
	return p.conn.Close(context.TODO())
}

func (p PostgresContext) IndexTransaction(txResult modules.TxResult) error {
	return p.txIndexer.Index(txResult)
}

func (p *PostgresContext) resetContext() (err error) {
	if p == nil {
		return nil
	}

	tx := p.GetTx()
	if tx == nil {
		return nil
	}

	conn := tx.Conn()
	if conn == nil {
		return nil
	}

	if !conn.IsClosed() {
		if err := conn.Close(context.TODO()); err != nil {
			return err
		}
	}

	p.conn = nil
	p.tx = nil
	p.blockHash = ""
	p.quorumCert = nil
	p.proposerAddr = nil
	p.blockTxs = nil

	return err
}
