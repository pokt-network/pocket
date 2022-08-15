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
