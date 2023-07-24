package cache

// TODO: add a TTL for cached sessions, since we know the sessions' length
import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/rpc"
)

var errSessionNotFound = errors.New("session not found in cache")

// SessionCache defines the set of methods used to interact with the client-side session cache
type SessionCache interface {
	Get(appAddr, chain string) (*rpc.Session, error)
	Set(session *rpc.Session) error
	Stop() error
}

// sessionCache stores and retrieves sessions for application+relaychain pairs
//
//	It uses a key-value store as backing storage
type sessionCache struct {
	// store is the local store for cached sessions
	store kvstore.KVStore
}

// Create returns a session cache backed by a kvstore using the provided database path.
func Create(databasePath string) (SessionCache, error) {
	store, err := kvstore.NewKVStore(databasePath)
	if err != nil {
		return nil, fmt.Errorf("Error initializing key-value store using path %s: %w", databasePath, err)
	}

	return &sessionCache{
		store: store,
	}, nil
}

// Get returns the cached session, if found, for an app+chain combination.
// The caller is responsible to verify that the returned session is valid for the current block height.
// Get is NOT safe to use concurrently
// DISCUSS: do we need concurrency here?
func (s *sessionCache) Get(appAddr, chain string) (*rpc.Session, error) {
	key := sessionKey(appAddr, chain)
	bz, err := s.store.Get(key)
	if err != nil {
		return nil, fmt.Errorf("error getting session from the store: %s %w", err.Error(), errSessionNotFound)
	}

	var session rpc.Session
	if err := json.Unmarshal(bz, &session); err != nil {
		return nil, fmt.Errorf("error unmarshalling session from store: %w", err)
	}

	return &session, nil
}

// Set stores the provided session in the cache with the key being the app+chain combination.
// For each app+chain combination, a single session will be stored. Subsequent calls to Set will overwrite the entry for the provided app and chain.
// Set is NOT safe to use concurrently
func (s *sessionCache) Set(session *rpc.Session) error {
	bz, err := json.Marshal(*session)
	if err != nil {
		return fmt.Errorf("error marshalling session for app: %s, chain: %s, session height: %d: %w", session.Application.Address, session.Chain, session.SessionHeight, err)
	}

	key := sessionKey(session.Application.Address, session.Chain)
	if err := s.store.Set(key, bz); err != nil {
		return fmt.Errorf("error storing session for app: %s, chain: %s, session height: %d in the cache: %w", session.Application.Address, session.Chain, session.SessionHeight, err)
	}
	return nil
}

// Stop call stop on the backing store. No calls should be made to Get or Set after calling Stop.
func (s *sessionCache) Stop() error {
	return s.store.Stop()
}

// sessionKey returns a key to get/set a session, based on application's address and the relay chain.
//
//	The height is not used as part of the key, because for each app+chain combination only one session, i.e. the current one, is of interest.
func sessionKey(appAddr, chain string) []byte {
	return []byte(fmt.Sprintf("%s-%s", appAddr, chain))
}
