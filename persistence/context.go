package persistence

import (
	"log"
)

func (p PostgresContext) NewSavePoint(bytes []byte) error {
	log.Println("TODO: Block - NewSavePoint not implemented")
	return nil
}

func (p PostgresContext) RollbackToSavePoint(bytes []byte) error {
	log.Println("TODO: Block - RollbackToSavePoint not implemented")
	return nil
}

func (p PostgresContext) AppHash() ([]byte, error) {
	log.Println("TODO: Block - AppHash not implemented")
	return []byte("A real app hash, I am not"), nil
}

func (p PostgresContext) Reset() error {
	log.Println("TODO: Block - Reset not implemented")
	return nil
}

func (p PostgresContext) Commit() error {
	log.Println("TODO: We have not implemented postgres based commits - it happens automatically")
	// INVESTIGATE:
	// 2. The data has already been written to the postgres DB, so what should we do here? The idea I have is:
	// - Call commit on the utility context
	// - Utility context maintains list of transactions to be applied
	// - Create a protobuf with the transactions -> serialized -> insert in the keystore
	return nil
}

func (p PostgresContext) Release() {
	log.Println("TODO:Block - Release not implemented")
}
