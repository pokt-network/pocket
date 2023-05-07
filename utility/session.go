package utility

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"

	"golang.org/x/exp/slices"

	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility/types"
)

const (
	NodeIsNotServicerErr = "node is not a servicer"
)

// TODO: Implement this
func (u *utilityModule) HandleRelay(relay *coreTypes.Relay) (*coreTypes.RelayResponse, error) {

	if u.servicer == nil {
		return nil, fmt.Errorf(NodeIsNotServicerErr)
	}

	return &coreTypes.RelayResponse{
		Payload:           "😎",
		ServicerSignature: "🪧",
	}, nil
}

// TODO: Implement this
func (u *utilityModule) HandleChallenge(challenge *coreTypes.Challenge) (*coreTypes.ChallengeResponse, error) {
	// References: https://github.com/pokt-network/pocket/pull/430
	//             https://forum.pokt.network/t/client-side-validation/148
	return nil, nil
}

// GetSession implements of the exposed `UtilityModule.GetSession` function
// TECHDEBT(#519): Add custom error types depending on the type of issue that occurred and assert on them in the unit tests.
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
		session:     session,
		blockHeight: height,
		readCtx:     readCtx,
	}

	if err := sessionHydrator.hydrateSessionMetadata(); err != nil {
		return nil, fmt.Errorf("failed to hydrate session metadata: %w", err)
	}

	if err := sessionHydrator.hydrateSessionApplication(appAddr); err != nil {
		return nil, fmt.Errorf("failed to hydrate session application: %w", err)
	}

	if err := sessionHydrator.validateApplicationSession(); err != nil {
		return nil, fmt.Errorf("failed to validate application session: %w", err)
	}

	if err := sessionHydrator.hydrateSessionID(); err != nil {
		return nil, fmt.Errorf("failed to hydrate session ID: %w", err)
	}

	if err := sessionHydrator.hydrateSessionServicers(); err != nil {
		return nil, fmt.Errorf("failed to hydrate session servicers: %w", err)
	}

	if err := sessionHydrator.hydrateSessionFishermen(); err != nil {
		return nil, fmt.Errorf("failed to hydrate session fishermen: %w", err)
	}

	return sessionHydrator.session, nil
}

type sessionHydrator struct {
	logger modules.Logger

	// The session being hydrated and returned
	session *coreTypes.Session

	// The height at which the request is being made to get session information
	blockHeight int64

	// Caches a readCtx to avoid draining too many connections to the database
	readCtx modules.PersistenceReadContext

	// A redundant helper that maintains a hex decoded copy of `session.Id` used for session hydration
	sessionIdBz []byte
}

// hydrateSessionMetadata hydrates the height at which the session started, its number, and the number of blocks per session
func (s *sessionHydrator) hydrateSessionMetadata() error {
	numBlocksPerSession, err := s.readCtx.GetIntParam(types.BlocksPerSessionParamName, s.blockHeight)
	if err != nil {
		return err
	}
	numBlocksAheadOfSession := s.blockHeight % int64(numBlocksPerSession)

	s.session.NumSessionBlocks = int64(numBlocksPerSession)
	s.session.SessionNumber = int64(s.blockHeight / int64(numBlocksPerSession))
	s.session.SessionHeight = s.blockHeight - numBlocksAheadOfSession
	return nil
}

// hydrateSessionApplication hydrates the full Application actor based on the address provided
func (s *sessionHydrator) hydrateSessionApplication(appAddr string) error {
	// TECHDEBT(#706): We can remove this decoding process once we use `strings` instead of `[]byte` for addresses
	addr, err := hex.DecodeString(appAddr)
	if err != nil {
		return err
	}
	s.session.Application, err = s.readCtx.GetActor(coreTypes.ActorType_ACTOR_TYPE_APP, addr, s.session.SessionHeight)
	return err
}

// validateApplicationSession validates that the application can have a valid session for the provided relay chain and geo zone
func (s *sessionHydrator) validateApplicationSession() error {
	app := s.session.Application

	if !slices.Contains(app.Chains, s.session.RelayChain) {
		return fmt.Errorf("application %s does not stake for relay chain %s", app.Address, s.session.RelayChain)
	}

	if app.PausedHeight != -1 || app.UnstakingHeight != -1 {
		return fmt.Errorf("application %s is either unstaked or paused", app.Address)
	}

	// TODO(#697): Filter by geo-zone

	// INVESTIGATE: Consider what else we should validate for here (e.g. Application stake amount, etc.)

	return nil
}

// hydrateSessionID use both session and on-chain data to determine a unique session ID
func (s *sessionHydrator) hydrateSessionID() error {
	sessionHeightBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sessionHeightBz, uint64(s.session.SessionHeight))

	prevHashHeight := int64(math.Max(float64(s.session.SessionHeight)-1, 0))
	prevHash, err := s.readCtx.GetBlockHash(prevHashHeight)
	if err != nil {
		return err
	}
	prevHashBz, err := hex.DecodeString(prevHash)

	if err != nil {
		return err
	}
	appPubKeyBz := []byte(s.session.Application.PublicKey)
	relayChainBz := []byte(string(s.session.RelayChain))
	geoZoneBz := []byte(s.session.GeoZone)

	s.sessionIdBz = concat(sessionHeightBz, prevHashBz, geoZoneBz, relayChainBz, appPubKeyBz)
	s.session.Id = crypto.GetHashStringFromBytes(s.sessionIdBz)

	return nil
}

// hydrateSessionServicers finds the servicers that are staked at the session height and populates the session with them
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

	// OPTIMIZE: Consider updating the persistence module so a single SQL query can retrieve all of the actors at once.
	candidateServicers := make([]*coreTypes.Actor, 0)
	for _, servicer := range servicers {
		// Sanity check the servicer is not paused, jailed or unstaking
		if servicer.PausedHeight != -1 || servicer.UnstakingHeight != -1 {
			return fmt.Errorf("hydrateSessionServicers should not have encountered a paused or unstaking servicer: %s", servicer.Address)
		}

		// TECHDEBT(#697): Filter by geo-zone

		// OPTIMIZE: If `servicer.Chains` was a map[string]struct{}, we could eliminate `slices.Contains()`'s loop
		if slices.Contains(servicer.Chains, s.session.RelayChain) {
			candidateServicers = append(candidateServicers, servicer)
		}
	}

	s.session.Servicers = pseudoRandomSelection(candidateServicers, numServicers, s.sessionIdBz)
	return nil
}

// hydrateSessionFishermen finds the fishermen that are staked at the session height and populates the session with them
func (s *sessionHydrator) hydrateSessionFishermen() error {
	// number of fishermen per session at this height
	numFishermen, err := s.readCtx.GetIntParam(types.FishermanPerSessionParamName, s.session.SessionHeight)
	if err != nil {
		return err
	}

	// returns all the staked fishermen at this session height
	fishermen, err := s.readCtx.GetAllFishermen(s.session.SessionHeight)
	if err != nil {
		return err
	}

	// OPTIMIZE: Consider updating the persistence module so a single SQL query can retrieve all of the actors at once.
	candidateFishermen := make([]*coreTypes.Actor, 0)
	for _, fisher := range fishermen {
		// Sanity check the fisher is not paused, jailed or unstaking
		if fisher.PausedHeight != -1 || fisher.UnstakingHeight != -1 {
			return fmt.Errorf("hydrateSessionFishermen should not have encountered a paused or unstaking fisherman: %s", fisher.Address)
		}

		// TODO(#697): Filter by geo-zone

		// OPTIMIZE: If this was a map[string]struct{}, we could have avoided the loop
		if slices.Contains(fisher.Chains, s.session.RelayChain) {
			candidateFishermen = append(candidateFishermen, fisher)
		}
	}

	s.session.Fishermen = pseudoRandomSelection(candidateFishermen, numFishermen, s.sessionIdBz)
	return nil
}

// pseudoRandomSelection returns a random subset of the candidates.
// DECIDE: We are using a `Go` native implementation for a pseudo-random number generator. In order
// for it to be language agnostic, a general purpose algorithm MUST be used.
func pseudoRandomSelection(candidates []*coreTypes.Actor, numTarget int, sessionId []byte) []*coreTypes.Actor {
	// If there aren't enough candidates, return all of them
	if numTarget > len(candidates) {
		logger.Global.Warn().Msgf("pseudoRandomSelection: numTarget (%d) is greater than the number of candidates (%d)", numTarget, len(candidates))
		return candidates
	}

	// Take the first 8 bytes of sessionId to use as the seed
	// NB: There is specific reason why `BigEndian` was chosen over `LittleEndian` in this specific context.
	seed := int64(binary.BigEndian.Uint64(crypto.SHA3Hash(sessionId)[:8]))

	// Retrieve the indices for the candidates
	actors := make([]*coreTypes.Actor, 0)
	uniqueIndices := uniqueRandomIndices(seed, int64(len(candidates)), int64(numTarget))
	for idx := range uniqueIndices {
		actors = append(actors, candidates[idx])
	}

	return actors
}

// OPTIMIZE: Postgres uses a `Twisted Mersenne Twister (TMT)` randomness algorithm.
// We could potentially look into changing everything into a single SQL query but
// would need to verify that it can be implemented in a platform agnostic way.

// uniqueRandomIndices returns a map of `numIndices` unique random numbers less than `maxIndex`
// seeded by `seed`.
// panics if `numIndicies > maxIndex` since that code path SHOULD never be executed.
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
