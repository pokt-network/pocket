package keybase_test

import (
	"github.com/dgraph-io/badger/v3"
	"log"
)

func main() {
	// Open the Badger database located in the given directory.
	// It will be created if it doesn't exist.
	// TODO: (team discuss) Where do we want to store key data?

	keyPath := "./key_data"
	db, err := badger.Open(badger.DefaultOptions(keyPath))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// MyCode...
}
