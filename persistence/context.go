package persistence

import (
	"context"
	"log"
)

func (p PostgresContext) NewSavePoint(bytes []byte) error {
	log.Println("TODO: NewSavePoint not implemented")
	return nil
}

func (p PostgresContext) RollbackToSavePoint(bytes []byte) error {
	log.Println("TODO: RollbackToSavePoint not fully implemented")
	return p.GetTx().Rollback(context.TODO())
}

func (p *PostgresContext) UpdateAppHash() ([]byte, error) {
	if err := p.updateStateHash(); err != nil {
		return nil, err
	}
	return p.currentStateHash, nil
}

// TODO_IN_THIS_COMMIT: Make sure that `prepareBlock`, `insertBlock`, and `storeBlock` are all atomic.
func (p PostgresContext) Commit(proposerAddr []byte, quorumCert []byte) error {
	log.Printf("About to commit context at height %d.\n", p.Height)

	// Create a persistence block proto
	block, err := p.prepareBlock(proposerAddr, quorumCert)
	if err != nil {
		return err
	}

	// Insert the block into the postgres DB
	if err := p.insertBlock(block); err != nil {
		return err
	}

	// Store block in the KV store
	if err := p.storeBlock(block); err != nil {
		return err
	}

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
	if err := p.conn.Close(ctx); err != nil {
		log.Println("[TODO][ERROR] Implement connection pooling. Error when closing DB connecting...", err)
	}
	return nil
}

func (p PostgresContext) Close() error {
	log.Printf("About to close context at height %d.\n", p.Height)

	return p.conn.Close(context.TODO())
}
