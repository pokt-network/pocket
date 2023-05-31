package trees_test

import (
	"testing"

	"github.com/pokt-network/pocket/persistence/trees"
	"github.com/stretchr/testify/assert"
)

func TestTreestoreUpdate(t *testing.T) {
	store, err := trees.NewtreeStore("")
	assert.NoError(t, err, "failed to get tree store")
	assert.NotNil(t, store, "got nil tree store")
}
