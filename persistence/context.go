package persistence

// TECHDEBT: Look into whether the receivers of `PostgresContext` could/should be pointers?

import (
	"context"
	"log"

	"github.com/pokt-network/pocket/shared/modules"
)

func (p PostgresContext) UpdateAppHash() ([]byte, error) {
	return []byte("TODO(#284): Implement this function."), nil
}

func (p PostgresContext) NewSavePoint(bytes []byte) error {
	log.Println("TODO: NewSavePoint not implemented")
	return nil
}

// TECHDEBT(#327): Guarantee atomicity betweens `prepareBlock`, `insertBlock` and `storeBlock` for save points & rollbacks.
func (p PostgresContext) RollbackToSavePoint(bytes []byte) error {
	log.Println("TODO: RollbackToSavePoint not fully implemented")
	return p.GetTx().Rollback(context.TODO())
}

func (p *PostgresContext) ComputeAppHash() ([]byte, error) {
	// IMPROVE(#361): Guarantee the integrity of the state
	//                Full details in the thread from the PR review: https://github.com/pokt-network/pocket/pull/285/files?show-viewed-files=true&file-filters%5B%5D=#r1033002640
	return p.updateMerkleTrees()
}

// TECHDEBT(#327): Make sure these operations are atomic
func (p PostgresContext) Commit(quorumCert []byte) error {
	log.Printf("About to commit block & context at height %d.\n", p.Height)

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

// INVESTIGATE(#361): Revisit if is used correctly in the context of the lifecycle of a persistenceContext and a utilityContext
func (p PostgresContext) IndexTransaction(txResult modules.TxResult) error {
	return p.txIndexer.Index(txResult)
}

func (p *PostgresContext) resetContext() (err error) {
	if p == nil {
		return nil
	}

	p.blockHash = ""
	p.quorumCert = nil
	p.proposerAddr = nil
	p.blockTxs = nil

	tx := p.GetTx()
	if p.tx == nil {
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

	return err
}
