package persistence

import (
	"fmt"
	"sync"

	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

type Node struct {
	ID int
}

type Coordinator struct {
	Stores []kvstore.KVStore
}

// Tx wraps a Block to apply it to the persistence layer
type Tx struct {
	Block *coreTypes.Block
}

// Commit applies a transaction to the three underlying stores of the persistence layer
func (c *Coordinator) Commit(txn *Tx) bool {
	var wg sync.WaitGroup
	results := make(chan bool, len(c.Stores))

	for _, store := range c.Stores {
		wg.Add(1)

		go func(node kvstore.KVStore) {
			defer wg.Done()

			// Prepare phase
			fmt.Printf("Node %v: received PREPARE for transaction\n", node)
			prepared := prepare(node)

			// Sending vote to the coordinator
			results <- prepared
		}(store)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// The coordinator collects votes
	for result := range results {
		if !result {
			// TODO: trigger a rollback here
			fmt.Println("TODO trigger a rollback on all module stores")
			return false
		}
	}

	// Commit phase
	for _, mod := range c.Stores {
		fmt.Printf("mod %v: COMMIT transaction\n", mod)
	}

	fmt.Println("Transaction commit")
	return true
}

// Apply applies an atomic commit to the persistence layer.
func Apply(stores []kvstore.KVStore, txn *Tx) error {
	coordinator := &Coordinator{
		Stores: stores,
	}
	if ok := coordinator.Commit(txn); !ok {
		return fmt.Errorf("ErrNoCommit: %v", txn)
	}
	return nil
}

func prepare(node kvstore.KVStore) bool {
	// prepare node by creating a savepoint here
	if err := node.Prepare(); err != nil {
		return false
	}
	return true
}
