package persistence

import (
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

type persistenceLocalContext struct {
	databasePath string
}

// INCOMPLETE: implement this
// StoreServiceRelay implements the PersistenceLocalContext interface
func (local *persistenceLocalContext) StoreServiceRelay(session *coreTypes.Session, appAddr string, relayDigest, relayReqResBytes []byte) error {
	return nil
}

// INCOMPLETE: implement this
// GetSessionTokensUsed implements the PersistenceLocalContext interface
func (local *persistenceLocalContext) GetSessionTokensUsed(*coreTypes.Session) (*big.Int, error) {
	return nil, nil
}

// INCOMPLETE: implement this
// Release implements the PersistenceLocalContext interface
func (m *persistenceLocalContext) Release() error {
	return nil
}
