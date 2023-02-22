package utility

// IMPORTANT: The interface and implementation defined in this file are for illustrative purposes only
// and need to be revisited before any implementation commences.

import (
	"encoding/binary"
	"encoding/hex"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility/types"
)

type RelayChain Identifier
type GeoZone Identifier

type Identifier interface {
	Name() string
	ID() string
	Bytes() []byte
}

type Session interface {
	NewSession(sessionHeight int64, blockHash string, geoZone GeoZone, relayChain RelayChain, application *coreTypes.Actor) (Session, types.Error)
	GetServicers() []*coreTypes.Actor // the Servicers providing Web3 to the application
	GetFishermen() []*coreTypes.Actor // the Fishermen monitoring the servicers
	GetApplication() *coreTypes.Actor // the Application consuming the web3 access
	GetRelayChain() RelayChain        // the chain identifier of the web3
	GetGeoZone() GeoZone              // the geo-location zone where all are registered
	GetSessionHeight() int64          // the block height when the session started
}

var _ Session = &session{}

type session struct {
	servicers     []*coreTypes.Actor
	fishermen     []*coreTypes.Actor
	application   *coreTypes.Actor
	relayChain    RelayChain
	geoZone       GeoZone
	blockHash     string
	key           []byte
	sessionHeight int64
}

func (s *session) NewSession(sessionHeight int64, blockHash string, geoZone GeoZone, relayChain RelayChain, application *coreTypes.Actor) (session Session, err types.Error) {
	s.sessionHeight = sessionHeight
	s.blockHash = blockHash
	s.geoZone = geoZone
	s.relayChain = relayChain
	s.application = application
	s.key, err = s.sessionKey()
	if err != nil {
		return
	}
	s.servicers = s.findClosestXServicers()
	s.fishermen = s.findClosestYFishermen()
	return s, nil
}

// use the seed information to determine a SHA3Hash that is used to find the closest N actors based
// by comparing the sessionKey with the actors' public key
func (s *session) sessionKey() ([]byte, types.Error) {
	sessionHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sessionHeightBytes, uint64(s.sessionHeight))
	blockHashBz, err := hex.DecodeString(s.blockHash)
	if err != nil {
		return nil, types.ErrHexDecodeFromString(err)
	}
	appPubKey, err := crypto.NewPublicKey(s.application.GetPublicKey())
	if err != nil {
		return nil, types.ErrNewPublicKeyFromBytes(err)
	}
	return concat(sessionHeightBytes, blockHashBz, s.geoZone.Bytes(), s.relayChain.Bytes(), appPubKey.Bytes()), nil
}

// uses the current 'world state' to determine the servicers in the session
// 1) get an ordered list of the public keys of servicers who are:
//   - actively staked
//   - staked within geo-zone
//   - staked for relay-chain
//
// 2) calls `pseudoRandomSelection(servicers, numberOfNodesPerSession)`
func (s *session) findClosestXServicers() []*coreTypes.Actor {
	// IMPORTANT:
	// THIS IS A DEMONSTRABLE FUNCTION THAT WILL NOT BE IMPLEMENTED AS SUCH
	// IT EXISTS IN THIS COMMIT PURELY TO COMMUNICATE THE EXPECTED BEHAVIOR
	return nil
}

// uses the current 'world state' to determine the fishermen in the session
// 1) get an ordered list of the public keys of fishermen who are:
//   - actively staked
//   - staked within geo-zone
//   - staked for relay-chain
//
// 2) calls `pseudoRandomSelection(fishermen, numberOfFishPerSession)`
func (s *session) findClosestYFishermen() []*coreTypes.Actor {
	// IMPORTANT:
	// THIS IS A DEMONSTRABLE FUNCTION THAT WILL NOT BE IMPLEMENTED AS SUCH
	// IT EXISTS IN THIS COMMIT PURELY TO COMMUNICATE THE EXPECTED BEHAVIOR
	return nil
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
func (s *session) pseudoRandomSelection(orderedListOfPublicKeys []string, numberOfActorsInSession int) []*coreTypes.Actor {
	// IMPORTANT:
	// THIS IS A DEMONSTRABLE FUNCTION THAT WILL NOT BE IMPLEMENTED AS SUCH
	// IT EXISTS IN THIS COMMIT PURELY TO COMMUNICATE THE EXPECTED BEHAVIOR
	return nil
}

func (s *session) GetServicers() []*coreTypes.Actor {
	return s.servicers
}

func (s *session) GetFishermen() []*coreTypes.Actor {
	return s.fishermen
}

func (s *session) GetApplication() *coreTypes.Actor {
	return s.application
}

func (s *session) GetRelayChain() RelayChain {
	return s.relayChain
}

func (s *session) GetGeoZone() GeoZone {
	return s.geoZone
}

func (s *session) GetSessionHeight() int64 {
	return s.sessionHeight
}

func concat(b ...[]byte) (result []byte) {
	for _, bz := range b {
		result = append(result, bz...)
	}
	return result
}
