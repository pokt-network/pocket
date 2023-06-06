package servicer

import (
	"crypto"
	"errors"
	"fmt"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"golang.org/x/exp/slices"
)

var (
	errValidateBlockHeight = errors.New("relay failed block height validation")
	errValidateRelayMeta   = errors.New("relay failed metadata validation")

	_ modules.ServicerModule = &servicer{}
)

const (
	ServicerModuleName = "servicer"
)

type servicer struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule

	logger *modules.Logger
	config *configs.ServicerConfig
}

var (
	_ modules.ServicerModule = &servicer{}
)

func CreateServicer(bus modules.Bus, options ...modules.ModuleOption) (modules.ServicerModule, error) {
	m, err := new(servicer).Create(bus, options...)
	if err != nil {
		return nil, err
	}
	return m.(modules.ServicerModule), nil
}

func (*servicer) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	s := &servicer{}

	for _, option := range options {
		option(s)
	}

	bus.RegisterModule(s)

	s.logger = logger.Global.CreateLoggerForModule(s.GetModuleName())

	cfg := bus.GetRuntimeMgr().GetConfig()
	s.config = cfg.Servicer

	return s, nil
}

// TODO: implement this function
func (s *servicer) Start() error {
	s.logger.Info().Msg("🧬 Servicer module started 🧬")
	return nil
}

func (s *servicer) Stop() error {
	s.logger.Info().Msg("🧬 Servicer module stopped 🧬")
	return nil
}

func (s *servicer) GetModuleName() string {
	return ServicerModuleName
}

// HandleRelay processes a relay after performing validation.
// It also updates the servicer's internal state to keep track of served relays.
func (s *servicer) HandleRelay(relay *coreTypes.Relay) (*coreTypes.RelayResponse, error) {
	if relay == nil {
		return nil, fmt.Errorf("cannot serve nil relay")
	}

	if err := s.admitRelay(relay); err != nil {
		return nil, fmt.Errorf("Error admitting relay: %w", err)
	}

	// TODO: implement Persist Relay
	// TODO: implement execution
	// TODO: implement state maintenance
	// TODO: validate the response from the node?
	// TODO: (QUESTION) Should we persist SignedRPC?
	return nil, nil
}

// validateRelayMeta ensures the relay metadata is valid for being handled by the servicer
// REFACTOR: move the meta-specific validation to a Validator method on RelayMeta struct
func (s servicer) validateRelayMeta(meta *coreTypes.RelayMeta, currentHeight int64) error {
	if meta == nil {
		return fmt.Errorf("empty relay metadata")
	}

	if meta.RelayChain == nil {
		return fmt.Errorf("relay chain unspecified")
	}

	// TODO: supported chains: needs to be crossed-checked with the world state from the persistence layer
	if err := s.validateRelayChainSupport(meta.RelayChain, currentHeight); err != nil {
		return fmt.Errorf("validation of support for relay chain %s failed: %w", meta.RelayChain.Id, err)
	}

	return nil
}

func (s servicer) validateRelayChainSupport(relayChain *coreTypes.Identifiable, currentHeight int64) error {
	if !slices.Contains(s.config.Chains, relayChain.Id) {
		return fmt.Errorf("chain %s not supported by servicer %s configuration", relayChain.Id, s.config.Address)
	}

	// DISCUSS: either update NewReadContext to take a uint64, or the GetCurrentHeight to return an int64.
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(currentHeight)
	if err != nil {
		return fmt.Errorf("error getting persistence context at height %d: %w", currentHeight, err)
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	// DISCUSS: should we update the GetServicer signature to take a string instead?
	servicer, err := readCtx.GetServicer([]byte(s.config.Address), currentHeight)
	if err != nil {
		return fmt.Errorf("error reading servicer from persistence: %w", err)
	}

	if !slices.Contains(servicer.Chains, relayChain.Id) {
		return fmt.Errorf("chain %s not supported by servicer %s configuration fetched from persistence", relayChain.Id, s.config.Address)
	}

	return nil
}

// TODO: implement
// validateApplication makes sure the application has not received more relays than allocated in the current session.
func (s servicer) validateApplication(meta *coreTypes.RelayMeta, session *coreTypes.Session) error {
	/*
		// if maxRelaysPerSession, overServiced := calculateAppSessionTokens(); overServiced {
			return fmt.Errorf("application %s has exceeded its allocated relays %d for the session %d", meta.ApplicationPublicKey, maxRelaysPerSession)
		}
	*/
	return nil
}

// validateServicer makes sure the servicer is A) active in the current session, and B) has not served more than its allocated relays for the session
func (s servicer) validateServicer(meta *coreTypes.RelayMeta, session *coreTypes.Session) error {
	if meta.ServicerPublicKey != s.config.PublicKey {
		return fmt.Errorf("relay servicer key %s does not match this servicer instance %s", meta.ServicerPublicKey, s.config.PublicKey)
	}

	var found bool
	for _, servicer := range session.Servicers {
		if servicer != nil && servicer.PublicKey == meta.ServicerPublicKey {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("relay servicer key %s not found in session %d with %d servicers", meta.ServicerPublicKey, session.SessionNumber, len(session.Servicers))
	}

	// TODO: implement isServicerMaxedOut
	return nil
}

// admitRelay decides whether the relay should be served
func (s servicer) admitRelay(relay *coreTypes.Relay) error {
	// TODO: utility module should initialize the servicer (if this module instance is a servicer)
	const errPrefix = "Error admitting relay"

	if relay == nil {
		return fmt.Errorf("%s: relay is nil", errPrefix)
	}

	height := s.GetBus().GetConsensusModule().CurrentHeight()
	if err := s.validateRelayMeta(relay.Meta, int64(height)); err != nil {
		return fmt.Errorf("%w: %s", errValidateRelayMeta, err.Error())
	}

	// TODO: update the CLI to include ApplicationAddress(or Application Public Key) in the RelayMeta
	session, err := s.GetBus().GetUtilityModule().GetSession(relay.Meta.ApplicationAddress, int64(height), relay.Meta.RelayChain.Id, relay.Meta.GeoZone.Id)
	if err != nil {
		return fmt.Errorf("%s: failed to get a session for height %d for relay meta %s: %w", errPrefix, height, relay.Meta, err)
	}

	// TODO: (REFACTOR) use a loop to run all validators: would also remove the need for passing the session around
	if err := validateRelayBlockHeight(relay.Meta, session); err != nil {
		return fmt.Errorf("%w: %s", errValidateBlockHeight, err.Error())
	}

	if err := s.validateApplication(relay.Meta, session); err != nil {
		return fmt.Errorf("%s: relay failed application validation: %w", errPrefix, err)
	}

	if err := s.validateServicer(relay.Meta, session); err != nil {
		return fmt.Errorf("%s: relay failed servicer instance validation: %w", errPrefix, err)
	}

	return nil
}

// IMPROVE: Add session height tolerance to account for session rollovers
func validateRelayBlockHeight(relayMeta *coreTypes.RelayMeta, session *coreTypes.Session) error {
	sessionStartingBlock := session.SessionNumber * session.NumSessionBlocks
	sessionLastBlock := sessionStartingBlock + session.SessionHeight

	if relayMeta.BlockHeight >= sessionStartingBlock && relayMeta.BlockHeight <= sessionLastBlock {
		return nil
	}

	return fmt.Errorf("relay block height %d not within session ID %s starting block %d and last block %d",
		relayMeta.BlockHeight,
		session.Id,
		sessionStartingBlock,
		sessionLastBlock)
}

// TECHDEBT: These structures were copied as placeholders from v0 and need to be updated to reflect changes in v1
// TODO: remove: use coreTypes.Relay instead
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
func (r *relay) Validate() coreTypes.Error {

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
func (r *relay) Store() coreTypes.Error {

	// marshal relay object into protoBytes

	// calculate the hashOf(protoBytes) <needed for volume tracking>

	// persist relay object, indexing under session

	return nil
}

// Execute a submitted relay by a client after validation
func (r *relay) Execute() (RelayResponse, coreTypes.Error) {

	// retrieve the RelayChain url from the servicer's local configuration file

	// execute http request with the relay payload

	// format and digitally sign the response

	return nil, nil
}

// Get volume metric applicable relays from store
func (r *relay) ReapStoreForHashCollision(sessionBlockHeight int64, hashEndWith string) ([]Relay, coreTypes.Error) {

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
func (r *relay) ReportVolumeMetrics(fishermanServiceURL string, volumeRelays []Relay) coreTypes.Error {

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
