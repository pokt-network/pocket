package utility

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"

	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility/types"
)

// OPTIMIZE: Postgres uses a `Twisted Mersenne Twister (TMT)` randomness algorithm.
// We could potentially look into changing everything a single SQL query but need to
// make sure that it can be implemented in a platform agnostic way.

// sessionHydrator is an internal structure used to prepare a Session returned by `GetSession` below
type sessionHydrator struct {
	logger modules.Logger

	// The height of the request for which the session is being hydrated
	blockHeight int64

	// The session being hydrated and returned
	session *coreTypes.Session

	// Caches a readCtx to avoid draining to many connections to the database
	readCtx modules.PersistenceReadContext

	// A redundant helper that keeps a hex decoded copy of `session.Id`
	sessionIdBz []byte
}

// GetSession is an implementation of the exposed `UtilityModule.GetSession` function
func (m *utilityModule) GetSession(appAddr string, height int64, relayChain, geoZone string) (*coreTypes.Session, error) {
	persistenceModule := m.GetBus().GetPersistenceModule()
	readCtx, err := persistenceModule.NewReadContext(height)
	if err != nil {
		return nil, err
	}
	defer readCtx.Release()

	session := &coreTypes.Session{
		RelayChain: relayChain,
		GeoZone:    geoZone,
	}

	sessionHydrator := &sessionHydrator{
		logger:      m.logger.With().Str("source", "sessionHydrator").Logger(),
		blockHeight: height,
		session:     session,
		readCtx:     readCtx,
	}

	if err := sessionHydrator.hydrateSessionHeight(height); err != nil {
		return nil, err
	}

	if err := sessionHydrator.hydrateSessionApplication(appAddr); err != nil {
		return nil, err
	}

	if err := sessionHydrator.validateApplicationDispatch(); err != nil {
		return nil, err
	}

	if err := sessionHydrator.hydrateSessionId(); err != nil {
		return nil, err
	}

	if err := sessionHydrator.hydrateSessionServicers(); err != nil {
		return nil, err
	}

	if err := sessionHydrator.hydrateSessionFishermen(); err != nil {
		return nil, err
	}

	return sessionHydrator.session, nil
}

// hydrateSessionHeight returns the height at which the session started given the current block height
func (s *sessionHydrator) hydrateSessionHeight(blockHeight int64) error {
	numBlocksPerSession, err := s.readCtx.GetIntParam(types.BlocksPerSessionParamName, blockHeight)
	if err != nil {
		return err
	}
	s.session.NumSessionBlocks = int64(numBlocksPerSession)

	numBlocksAheadOfSession := blockHeight % int64(numBlocksPerSession)
	s.session.SessionNumber = int64(blockHeight / int64(numBlocksPerSession))
	if numBlocksAheadOfSession == 0 {
		s.session.SessionHeight = blockHeight
	} else {
		s.session.SessionHeight = blockHeight - numBlocksAheadOfSession
	}
	return nil
}

// use the seed information to determine a SHA3Hash that is used to find the closest N actors based
// by comparing the sessionKey with the actors' public key
func (s *sessionHydrator) hydrateSessionId() error {
	sessionHeightBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sessionHeightBz, uint64(s.session.SessionHeight))
	prevHashHeight := int64(math.Max(float64(s.session.SessionHeight)-1, 0))
	prevHash, err := s.readCtx.GetBlockHash(prevHashHeight)
	if err != nil {
		return err
	}
	prevHashBz, err := hex.DecodeString(prevHash)
	appPubKeyBz := []byte(s.session.Application.PublicKey)
	relayChainBz := []byte(string(s.session.RelayChain))
	geoZoneBz := []byte(s.session.GeoZone)
	s.sessionIdBz = concat(sessionHeightBz, prevHashBz, geoZoneBz, relayChainBz, appPubKeyBz)
	s.session.Id = crypto.GetHashStringFromBytes(s.sessionIdBz)
	return nil
}

// Uses the current 'world state' to determine the full application metadata based on its address at the current height
func (s *sessionHydrator) hydrateSessionApplication(appAddr string) error {
	// TECHDEBT: We can remove this decoding process once we use `strings` instead of `[]byte` for addresses
	addr, err := hex.DecodeString(appAddr)
	if err != nil {
		return err
	}
	s.session.Application, err = s.readCtx.GetActor(coreTypes.ActorType_ACTOR_TYPE_APP, addr, s.session.SessionHeight)
	return err
}

// Validate the the application can dispatch a session at the requested geo-zone and for the request relay chain
func (s *sessionHydrator) validateApplicationDispatch() error {
	// if s.session.Application.Chains {
	// TECHDEBT: We can remove this decoding process once we use `strings` instead of `[]byte` for addresses
	// addr, err := hex.DecodeString(s.session.Application.Address)
	// if err != nil {
	// 	return err
	// }
	// s.session.Application, err = s.readCtx.GetActor(coreTypes.ActorType_ACTOR_TYPE_APP, addr, s.session.SessionHeight)
	// return err
	return nil
}

// uses the current 'world state' to determine the servicers in the session
// 1) get an ordered list of the public keys of servicers who are:
//   - actively staked
//   - staked within geo-zone (or closest geo-zones)
//   - staked for relay-chain
//
// 2) calls `pseudoRandomSelection(servicers, numberOfNodesPerSession)`
func (s *sessionHydrator) hydrateSessionServicers() error {
	// number of servicers per session at this height
	numServicers, err := s.readCtx.GetIntParam(types.ServicersPerSessionParamName, s.session.SessionHeight)
	if err != nil {
		return err
	}

	// returns all the staked servicers at this session height
	servicers, err := s.readCtx.GetAllServicers(s.session.SessionHeight)
	if err != nil {
		return err
	}

	// OPTIMIZE: Update the persistence module to allow for querying for filtered servicers directly
	// Determine the servicers for this session
	candidateServicers := make([]*coreTypes.Actor, 0)
	for _, servicer := range servicers {
		// Sanity check the servicer is not paused or unstaking
		if !(servicer.PausedHeight == -1 && servicer.UnstakingHeight == -1) {
			return fmt.Errorf("hydrateSessionServicers should not have encountered a paused or unstaking servicer: %s", servicer.Address)
		}

		// TODO(#XXX): Add GeoZone filtering

		// OPTIMIZE: If this was a map, we could have avoided the loop over chains
		var chain string
		for _, chain = range servicer.Chains {
			if chain != s.session.RelayChain {
				chain = ""
				continue
			}
		}
		if chain != "" {
			candidateServicers = append(candidateServicers, servicer)
		}
	}

	s.session.Servicers = pseudoRandomSelection(candidateServicers, numServicers, s.sessionIdBz)
	return nil
}

// uses the current 'world state' to determine the fishermen in the session
// 1) get an ordered list of the public keys of fishermen who are:
//   - actively staked
//   - staked within geo-zone  (or closest geo-zones)
//   - staked for relay-chain
//
// 2) calls `pseudoRandomSelection(fishermen, numberOfFishPerSession)`
func (s *sessionHydrator) hydrateSessionFishermen() error {
	// number of fisherman per session at this height
	numFishermen, err := s.readCtx.GetIntParam(types.FishermanPerSessionParamName, s.session.SessionHeight)
	if err != nil {
		return err
	}

	// returns all the staked fisherman at this session height
	fishermen, err := s.readCtx.GetAllFishermen(s.session.SessionHeight)
	if err != nil {
		return err
	}

	// OPTIMIZE: Update the persistence module to allow for querying for filtered fishermen directly
	// Determine the fishermen for this session
	candidateFishermen := make([]*coreTypes.Actor, 0)
	for _, fisherman := range fishermen {
		// Sanity check the fisherman is not paused or unstaking
		if !(fisherman.PausedHeight == -1 && fisherman.UnstakingHeight == -1) {
			return fmt.Errorf("hydrateSessionFishermen should not have encountered a paused or unstaking fisherman: %s", fisherman.Address)
		}

		// TODO_IN_THIS_COMMIT: if sfisherman.GeoZone includes session.GeoZone

		// OPTIMIZE: If this was a map, we could have avoided the loop over chains
		var chain string
		for _, chain = range fisherman.Chains {
			if chain != s.session.RelayChain {
				chain = ""
				continue
			}
		}
		if chain != "" {
			candidateFishermen = append(candidateFishermen, fisherman)
		}
	}

	s.session.Fishermen = pseudoRandomSelection(candidateFishermen, numFishermen, s.sessionIdBz)
	return nil
}

// pseudoRandomSelection returns a random subset of the candidates.
// TECHDEBT: We are using a `Go` native implementation for a pseudo-random number generator. In order
// for it to be language agnostic, a general purpose algorithm needs ot be used.
func pseudoRandomSelection(candidates []*coreTypes.Actor, numTarget int, sessionId []byte) []*coreTypes.Actor {
	// If there aren't enough candidates, return all of them
	if numTarget > len(candidates) {
		logger.Global.Warn().Msgf("pseudoRandomSelection: numTarget (%d) is greater than the number of candidates (%d)", numTarget, len(candidates))
		return candidates
	}

	// Take the first 8 bytes of sessionId to use as the seed
	seed := int64(binary.BigEndian.Uint64(crypto.SHA3Hash(sessionId)[:8]))

	// Retrieve the indices for the candidates
	actors := make([]*coreTypes.Actor, 0)
	uniqueIndices := uniqueRandomIndices(seed, int64(len(candidates)), int64(numTarget))
	for idx := range uniqueIndices {
		actors = append(actors, candidates[idx])
	}

	return actors
}

// uniqueRandomIndices returns a map of `numIndices` unique random numbers less than `maxIndex`
// seeded by `seed`.
// NB: A map pointing to empty structs is used to simulate set behaviour.
func uniqueRandomIndices(seed, maxIndex, numIndices int64) map[int64]struct{} {
	// This should never happen
	if numIndices > maxIndex {
		panic(fmt.Sprintf("uniqueRandomIndices: numIndices (%d) is greater than maxIndex (%d)", numIndices, maxIndex))
	}

	// create a new random source with the seed
	randSrc := rand.NewSource(seed)

	// initialize a map to capture the indicesMap we'll return
	indicesMap := make(map[int64]struct{}, maxIndex)

	// The random source could potentially return duplicates, so while loop until we have enough unique indices
	for int64(len(indicesMap)) < numIndices {
		indicesMap[randSrc.Int63()%int64(maxIndex)] = struct{}{}
	}

	return indicesMap
}

func concat(b ...[]byte) (result []byte) {
	for _, bz := range b {
		result = append(result, bz...)
	}
	return result
}
