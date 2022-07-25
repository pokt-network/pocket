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
	log.Println("TODO: Reset not implemented")
	return nil
}

// HACK: The data has already been written to the postgres DB, so what should we do here? The idea I have is:
func (p PostgresContext) Commit() error {
	log.Println("TODO: We have not implemented postgres based commits - it happens automatically")
	return nil
}

func (p PostgresContext) Release() {
	log.Println("TODO:Block - Release not implemented")
}
