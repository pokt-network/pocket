package service

import (
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility/types"
)

type Relay interface {
	RelayPayload
	RelayMeta
}

type RelayPayload interface {
	GetData() string               // the actual data string for the external chain
	GetMethod() string             // the http CRUD method
	GetHTTPPath() string           // the HTTP Path
	GetHeaders() map[string]string // http headers
}

type RelayMeta interface {
	GetBlockHeight() int64 // the block height when the request is made
	GetServicerPublicKey() crypto.PublicKey
	GetRelayChain() RelayChain
	GetGeoZone() GeoZone
	GetToken() AAT
	GetSignature() string
}

type RelayResponse interface {
	Payload() string
	ServicerSignature() string
}

type RelayChain Identifiable
type GeoZone Identifiable

type AAT interface {
	GetVersion() string              // confirm a valid AAT version
	GetApplicationPublicKey() string // confirm the identity/signature of the app
	GetClientPublicKey() string      // confirm the identity/signature of the client
	GetApplicationSignature() string // confirm the application signed the token
}

type Identifiable interface {
	Name() string
	ID() string
}

var _ Relay = &relay{}

type relay struct{}

// Validate a submitted relay by a client before servicing
func (r *relay) Validate() types.Error {

	// validate payload

	// validate the metadata

	// ensure the RelayChain is supported locally

	// ensure session block height is current

	// get the session context

	// get the application object from the r.AAT()

	// get session node count from that session height

	// get maximum possible relays for the application

	// ensure not over serviced

	// generate the session from seed data

	// validate self against the session

	return nil
}

// Store a submitted relay by a client for volume tracking
func (r *relay) Store() types.Error {

	// marshal relay object into protoBytes

	// calculate the hashOf(protoBytes) <needed for volume tracking>

	// persist relay object, indexing under session

	return nil
}

// Execute a submitted relay by a client after validation
func (r *relay) Execute() (RelayResponse, types.Error) {

	// retrieve the RelayChain url from the servicer's local configuration file

	// execute http request with the relay payload

	// format and digitally sign the response

	return nil, nil
}

// Get volume metric applicable relays from store
func (r *relay) ReapStoreForHashCollision(sessionBlockHeight int64, hashEndWith string) ([]Relay, types.Error) {

	// Pull all relays whose hash collides with the revealed secret key
	// It's important to note, the secret key isn't revealed by the network until the session is over
	// to prevent volume based bias. The secret key is usually a pseudorandom selection using the block hash as a seed.
	// (See the session protocol)
	//
	// Demonstrable pseudocode below:
	//   `SELECT * from RELAY where HashOf(relay) ends with hashEndWith AND sessionBlockHeight=sessionBlockHeight`

	// This function also signifies deleting the non-volume-applicable Relays

	return nil, nil
}

// Report volume metric applicable relays to Fisherman
func (r *relay) ReportVolumeMetrics(fishermanServiceURL string, volumeRelays []Relay) types.Error {

	// Send all volume applicable relays to the assigned trusted Fisherman for
	// a proper verification of the volume completed. Send volumeRelays to fishermanServiceURL
	// through http.

	// NOTE: an alternative design is a 2 step, claim - proof lifecycle where the individual servicers
	// build a merkle sum index tree from all the relays, submits a root and subsequent merkle proof to the
	// network.
	//
	// Pros: Can report volume metrics directly to the chain in a trustless fashion
	// Cons: Large chain bloat, non-trivial compute requirement for creation of claim/proof transactions and trees,
	//       non-trivial compute requirement to process claim / proofs during ApplyBlock()

	return nil
}

func (r *relay) GetData() string                        { return "" }
func (r *relay) GetMethod() string                      { return "" }
func (r *relay) GetHTTPPath() string                    { return "" }
func (r *relay) GetHeaders() map[string]string          { return nil }
func (r *relay) GetBlockHeight() int64                  { return 0 }
func (r *relay) GetServicerPublicKey() crypto.PublicKey { return nil }
func (r *relay) GetRelayChain() RelayChain              { return nil }
func (r *relay) GetGeoZone() GeoZone                    { return nil }
func (r *relay) GetToken() AAT                          { return nil }
func (r *relay) GetSignature() string                   { return "" }
