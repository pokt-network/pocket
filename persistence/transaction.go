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
			err := store.Prepare(c.tx)

			// Sending vote to the coordinator
			if err != nil {
				results <- false
			} else {
				results <- true
			}
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
			return false
		}
	}

	// Commit phase
	for _, store := range c.stores {
		fmt.Printf("mod %v: COMMIT transaction\n", store)
		if err := store.Commit(); err != nil {
			c.rollback()
			return false
		}
	}

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
