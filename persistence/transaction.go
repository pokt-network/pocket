package persistence

import (
	"fmt"
	"sync"

	"github.com/pokt-network/pocket/shared/modules"
)

type Node struct {
	ID int
}

// StoreManager handles the atomic commit or rollback of a Tx to a set of Revertible Stores
type StoreManager struct {
	stores []modules.AtomicStore
	tx     modules.Tx
}

// Commit applies a transaction to the three underlying stores of the persistence layer
func (c *StoreManager) Commit() bool {
	var wg sync.WaitGroup
	results := make(chan bool, len(c.stores))

	for _, store := range c.stores {
		wg.Add(1)

		go func(store modules.AtomicStore) {
			defer wg.Done()

			// Prepare phase
			fmt.Printf("store %v: received PREPARE for transaction\n", store)
			prepared := prepare(store, c.tx)

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
			c.rollback()
			fmt.Println("TODO trigger a rollback on all module stores")
			return false
		}
	}

	// Commit phase
	for _, store := range c.stores {
		fmt.Printf("mod %v: COMMIT transaction\n", store)
		if err := store.Commit(); err != nil {
			// trigger coordinator rollback if any one commit fails
			c.rollback()
		}
	}

	fmt.Println("Transaction commit")
	return true
}

// rollback calls Rollback on each store in the StoreManager.
func (c *StoreManager) rollback() {
	for _, store := range c.stores {
		store.Rollback()
	}
}

// Apply applies an atomic commit to the persistence layer consisting of the given stores.
func Apply(stores []modules.AtomicStore, txn modules.Tx) error {
	sm := &StoreManager{
		stores: stores,
	}
	for _, r := range stores {
		if err := r.Prepare(txn); err != nil {
			sm.rollback()
			return err
		}
	}
	// attempt to commit
	if ok := sm.Commit(); !ok {
		sm.rollback()
		return fmt.Errorf("ErrNoCommit: %v", txn)
	}
	return nil
}

func prepare(node modules.AtomicStore, tx modules.Tx) bool {
	// prepare node by creating a savepoint here
	if err := node.Prepare(tx); err != nil {
		return false
	}
	return true
}
