package persistence

import (
	"context"
	"log"
)

func (p PostgresContext) UpdateAppHash() ([]byte, error) {
	panic("INTRODUCE(#284): Add this function in #284 per the interface changes in #252.")
}

// func (p PostgresContext) Commit(proposerAddr []byte, quorumCert []byte) error {
// 	panic("INTRODUCE(#284): Add this function in #284 per the interface changes in #252.")
// }

func (p PostgresContext) NewSavePoint(bytes []byte) error {
	log.Println("TODO: NewSavePoint not implemented")
	return nil
}

func (p PostgresContext) RollbackToSavePoint(bytes []byte) error {
	log.Println("TODO: RollbackToSavePoint not fully implemented")
	return p.GetTx().Rollback(context.TODO())
}

func (p PostgresContext) AppHash() ([]byte, error) {
	log.Println("TODO: AppHash not implemented")
	return []byte("A real app hash, I am not"), nil
}

func (p *PostgresContext) Reset() error {
	p.txResults = nil
	p.blockHash = ""
	p.quorumCertificate = nil
	p.proposerAddr = nil
	p.blockProtoBytes = nil
	p.blockTxs = nil
	return nil
}

func (p PostgresContext) Commit() error {
	log.Printf("About to commit context at height %d.\n", p.Height)

	ctx := context.TODO()
	if err := p.GetTx().Commit(context.TODO()); err != nil {
		return err
	}
	if err := p.StoreBlock(); err != nil {
		return err
	}
	if err := p.conn.Close(ctx); err != nil {
		log.Println("[TODO][ERROR] Implement connection pooling. Error when closing DB connecting...", err)
	}
	if err := p.Reset(); err != nil {
		return err
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
	if err := p.Reset(); err != nil {
		return err
	}
	return nil
}

func (p PostgresContext) Close() error {
	log.Printf("About to close context at height %d.\n", p.Height)

	return p.conn.Close(context.TODO())
}
