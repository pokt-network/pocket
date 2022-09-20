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
	return p.DB.Tx.Rollback(context.TODO())
}

func (p PostgresContext) UpdateAppHash() ([]byte, error) {
	if _, err := p.updateStateHash(); err != nil {
		return nil, err
	}
	return p.StateHash, nil
}

func (p PostgresContext) AppHash() ([]byte, error) {
	// log.Println("TODO: AppHash not implemented")
	// return []byte("A real app hash, I am not"), n
	return p.StateHash, nil
}

func (p PostgresContext) Reset() error {
	panic("TODO: PostgresContext Reset not implemented")
}

func (p PostgresContext) Commit() error {
	log.Printf("About to commit context at height %d.\n", p.Height)

	// HACK: The data has already been written to the postgres DB, so what should we do here? The idea I have is:
	// if _, err := p.updateStateHash(); err != nil {
	// 	return err
	// }

	ctx := context.TODO()
	if err := p.DB.Tx.Commit(context.TODO()); err != nil {
		return err
	}
	if err := p.DB.conn.Close(ctx); err != nil {
		log.Println("[TODO][ERROR] Implement connection pooling. Error when closing DB connecting...", err)

	}
	return nil
}

func (p PostgresContext) Release() error {
	log.Printf("About to release context at height %d.\n", p.Height)

	ctx := context.TODO()
	if err := p.DB.Tx.Rollback(ctx); err != nil {
		return err
	}
	if err := p.DB.conn.Close(ctx); err != nil {
		log.Println("[TODO][ERROR] Implement connection pooling. Error when closing DB connecting...", err)
	}
	return nil
}

func (p PostgresContext) Close() error {
	log.Printf("About to close context at height %d.\n", p.Height)

	return p.DB.conn.Close(context.TODO())
}
