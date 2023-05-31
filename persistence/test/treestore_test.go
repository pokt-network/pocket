package test

import (
	"testing"

	"github.com/pokt-network/pocket/persistence/trees"
	"github.com/stretchr/testify/assert"
)

func TestTreestoreUpdate(t *testing.T) {
	pctx := NewTestPostgresContext(t, 0)
	assert.NotNil(t, pctx, "failed to get new test postgres context")
	store, err := trees.NewtreeStore("")
	assert.NoError(t, err, "failed to get tree store")
	assert.NotNil(t, store, "got nil tree store")
}
