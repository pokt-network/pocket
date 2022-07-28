package persistence

import (
	"log"
)

func (p PostgresContext) NewSavePoint(bytes []byte) error {
	log.Println("TODO: NewSavePoint not implemented")
	return nil
}

func (p PostgresContext) RollbackToSavePoint(bytes []byte) error {
	log.Println("TODO: RollbackToSavePoint not implemented")
	return nil
}

func (p PostgresContext) AppHash() ([]byte, error) {
	log.Println("TODO: AppHash not implemented")
	return []byte("A real app hash, I am not"), nil
}

func (p PostgresContext) Reset() error {
	panic("TODO: PostgresContext Reset not implemented")
}

func (p PostgresContext) Commit() error {
	// HACK: The data has already been written to the postgres DB, so what should we do here? The idea I have is:
	log.Println("TODO: Postgres context commit is currently a NOOP")
	return nil
}

func (p PostgresContext) Release() {
	if err := p.ContextStore.Stop(); err != nil {
		log.Printf("[ERROR] stopping postgres context store: %s", err)
	}
}
