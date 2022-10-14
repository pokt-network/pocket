package utility

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility/types"
)

type Session interface {
	NewSession(sessionHeight int64, blockHash string, geoZone GeoZone, relayChain RelayChain, application modules.Actor) (Session, types.Error)
	GetServiceNodes() []modules.Actor // the ServiceNodes providing Web3 to the application
	GetFishermen() []modules.Actor    // the Fishermen monitoring the serviceNodes
	GetApplication() modules.Actor    // the Application consuming the web3 access
	GetRelayChain() RelayChain        // the chain identifier of the web3
	GetGeoZone() GeoZone              // the geolocation zone where all are registered
	GetSessionHeight() int64          // the block height when the session started
}

type RelayChain Identifier
type GeoZone Identifier

type Identifier interface {
	Name() string
	ID() string
	Bytes() []byte
}

var _ Session = &session{}

type session struct {
	serviceNodes  []modules.Actor
	fishermen     []modules.Actor
	application   modules.Actor
	relayChain    RelayChain
	geoZone       GeoZone
	blockHash     string
	key           []byte
	sessionHeight int64
}

func (s *session) NewSession(sessionHeight int64, blockHash string, geoZone GeoZone, relayChain RelayChain, application modules.Actor) (session Session, err types.Error) {
	s.sessionHeight = sessionHeight
	s.blockHash = blockHash
	s.geoZone = geoZone
	s.relayChain = relayChain
	s.application = application
	s.key, err = s.sessionKey()
	if err != nil {
		return
	}
	s.serviceNodes = s.findClosestXServiceNodes()
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
	return s.concat(sessionHeightBytes, blockHashBz, s.geoZone.Bytes(), s.relayChain.Bytes(), appPubKey.Bytes()), nil
}

// uses the current 'world state' to determine the service nodes in the session
// 1) get an ordered list of the public keys of service nodes who are:
//    - actively staked
//    - staked within geo-zone
//    - staked for relay-chain
// 2) calls `pseudoRandomSelection(serviceNodes, numberOfNodesPerSession)`
func (s *session) findClosestXServiceNodes() []modules.Actor {
	// TODO (@andrewnguyen22) implement me
	return nil
}

// uses the current 'world state' to determine the fishermen in the session
// 1) get an ordered list of the public keys of fishermen who are:
//    - actively staked
//    - staked within geo-zone
//    - staked for relay-chain
// 2) calls `pseudoRandomSelection(fishermen, numberOfFishPerSession)`
func (s *session) findClosestYFishermen() []modules.Actor {
	// TODO (@andrewnguyen22) implement me
	return nil
}

// 1) passed an ordered list of the public keys of actors and number of nodes
// 2) pseudo-insert the session `key` string into the list and find the first actor directly below
// 3) newKey = Hash( key + actor1PublicKey )
// 4) repeat steps 2 and 3 until all N actor are found
// FAQ:
// Q) why do we hash to find a newKey between every actor selection?
// A) pseudo-random selection only works if each iteration is re-randomized
//    or it would be subject to lexicographical proximity bias attacks
func (s *session) pseudoRandomSelection(orderedListOfPublicKeys []string, numberOfActorsInSession int) []modules.Actor {
	// TODO (@andrewnguyen22) implement me
	return nil
}

func (s *session) GetServiceNodes() []modules.Actor {
	return s.serviceNodes
}

func (s *session) GetFishermen() []modules.Actor {
	return s.fishermen
}

func (s *session) GetApplication() modules.Actor {
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

func (s *session) concat(b ...[]byte) (result []byte) {
	for _, bz := range b {
		result = append(result, bz...)
	}
	return result
}
