package utility

// IMPORTANT: The interface and implementation defined in this file are for illustrative purposes only
// and need to be revisited before any implementation commences.

import (
	"encoding/binary"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility/types"
)

// TODO: When implementing please review if block height tolerance (+,-1) is included in the session protocol: pokt-network/pocket-core#1464 CC @Olshansk

// REFACTOR: Move these into `utility/types` and consider creating an enum
type RelayChain string
type GeoZone string

type Session interface {
	NewSession(sessionHeight int64, relayChain RelayChain, geoZone GeoZone, application *coreTypes.Actor) (Session, types.Error)
	GetSessionID() []byte             // the identifier of the dispatched session
	GetSessionHeight() int64          // the block height when the session started
	GetRelayChain() RelayChain        // the web3 chain identifier
	GetGeoZone() GeoZone              // the geo-location zone where the application is intending to operate during the session
	GetApplication() *coreTypes.Actor // the Application consuming the web3 access
	GetServicers() []*coreTypes.Actor // the Servicers providing Web3 to the application
	GetFishermen() []*coreTypes.Actor // the Fishermen monitoring the servicers
}

var _ Session = &session{}

type session struct {
	sessionId   []byte
	height      int64
	relayChain  RelayChain
	geoZone     GeoZone
	application *coreTypes.Actor
	servicers   []*coreTypes.Actor
	fishermen   []*coreTypes.Actor
}

func (*session) NewSession(
	sessionHeight int64,
	relayChain RelayChain,
	geoZone GeoZone,
	application *coreTypes.Actor,
) (Session, types.Error) {
	s := &session{
		height:      sessionHeight,
		relayChain:  relayChain,
		geoZone:     geoZone,
		application: application,
	}

	// TODO: make these configurable or based on governance params
	numServicers := 1
	numFisherman := 1

	var err types.Error
	if s.servicers, err = s.selectSessionServicers(numServicers); err != nil {
		return nil, err
	}
	if s.fishermen, err = s.selectSessionFishermen(numFisherman); err != nil {
		return nil, err
	}
	if s.sessionId, err = s.getSessionId(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *session) GetSessionID() []byte {
	return s.sessionId
}

func (s *session) GetSessionHeight() int64 {
	return s.height
}

func (s *session) GetRelayChain() RelayChain {
	return s.relayChain
}

func (s *session) GetGeoZone() GeoZone {
	return s.geoZone
}

func (s *session) GetApplication() *coreTypes.Actor {
	return s.application
}

func (s *session) GetFishermen() []*coreTypes.Actor {
	return s.fishermen
}

func (s *session) GetServicers() []*coreTypes.Actor {
	return s.servicers
}

// use the seed information to determine a SHA3Hash that is used to find the closest N actors based
// by comparing the sessionKey with the actors' public key
func (s *session) getSessionId() ([]byte, types.Error) {
	sessionHeightBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sessionHeightBz, uint64(s.height))

	blockHashBz := []byte("get block hash bytes at s.sessionHeight from persistence module")

	appPubKey, err := crypto.NewPublicKey(s.application.GetPublicKey())
	if err != nil {
		return nil, types.ErrNewPublicKeyFromBytes(err)
	}

	return concat(sessionHeightBz, blockHashBz, []byte(s.geoZone), []byte(s.relayChain), appPubKey.Bytes()), nil
}

// uses the current 'world state' to determine the servicers in the session
// 1) get an ordered list of the public keys of servicers who are:
//   - actively staked
//   - staked within geo-zone (or closest geo-zones)
//   - staked for relay-chain
//
// 2) calls `pseudoRandomSelection(servicers, numberOfNodesPerSession)`
func (s *session) selectSessionServicers(numServicers int) ([]*coreTypes.Actor, types.Error) {
	// IMPORTANT: This function is for behaviour illustrative purposes only and implementation may differ.
	return nil, nil
}

// uses the current 'world state' to determine the fishermen in the session
// 1) get an ordered list of the public keys of fishermen who are:
//   - actively staked
//   - staked within geo-zone  (or closest geo-zones)
//   - staked for relay-chain
//
// 2) calls `pseudoRandomSelection(fishermen, numberOfFishPerSession)`
func (s *session) selectSessionFishermen(numFishermen int) ([]*coreTypes.Actor, types.Error) {
	// IMPORTANT: This function is for behaviour illustrative purposes only and implementation may differ.
	return nil, nil
}

// 1) passed an ordered list of the public keys of actors and number of nodes
// 2) pseudo-insert the session `key` string into the list and find the first actor directly below
// 3) newKey = Hash( key + actor1PublicKey )
// 4) repeat steps 2 and 3 until all N actor are found
// FAQ:
// Q) why do we hash to find a newKey between every actor selection?
// A) pseudo-random selection only works if each iteration is re-randomized
//
//	or it would be subject to lexicographical proximity bias attacks
//
//nolint:unused // This is a demonstratable function
func (s *session) pseudoRandomSelection(orderedListOfPublicKeys []string, numActorsToSelect int) []*coreTypes.Actor {
	// IMPORTANT: This function is for behaviour illustrative purposes only and implementation may differ.
	return nil
}

func concat(b ...[]byte) (result []byte) {
	for _, bz := range b {
		result = append(result, bz...)
	}
	return result
}
