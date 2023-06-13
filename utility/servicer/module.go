package servicer

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"golang.org/x/exp/slices"
)

var (
	errValidateBlockHeight = errors.New("relay failed block height validation")
	errValidateRelayMeta   = errors.New("relay failed metadata validation")
	errValidateServicer    = errors.New("relay failed servicer validation")
	errValidateApplication = errors.New("relay failed application validation")

	_ modules.ServicerModule = &servicer{}
)

const (
	ServicerModuleName = "servicer"
)

// sessionTokens is used to cache the original number of tokens available
// during a specific session: it is used as the value for a map with keys being applications' public keys
type sessionTokens struct {
	sessionNumber          int64
	originalCountAvailable *big.Int
}

type servicer struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule

	logger *modules.Logger
	config *configs.ServicerConfig

	// This lock is needed to allow multiple GO routines update the totalTokens cache as part of serving relays
	rwlock sync.RWMutex
	// totalTokens is a mapping from application public keys to session metadata to keep track of session tokens
	// INVESTIGATE: considering the computational complexity, should we skip caching this value?
	totalTokens map[string]*sessionTokens
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

	response, err := s.executeRelay(relay)
	if err != nil {
		return nil, fmt.Errorf("Error executing relay: %w", err)
	}

	// TODO(M6): Look into data integrity checks and response validation.

	relayDigest, relayReqResBytes, shouldStore, err := s.isRelayVolumeApplicable(relay, response)
	if err != nil {
		return nil, fmt.Errorf("Error calculating relay service digest: %w", err)
	}
	if !shouldStore {
		return response, nil
	}

	session, err := s.getSession(relay)
	if err != nil {
		return nil, err
	}

	localCtx, err := s.GetBus().GetPersistenceModule().NewLocalContext()
	if err != nil {
		return nil, fmt.Errorf("Error getting a local context to update token usage for application %s: %w", relay.Meta.ApplicationAddress, err)
	}

	if err := localCtx.StoreServiceRelay(session, relay.Meta.ApplicationAddress, relayDigest, relayReqResBytes); err != nil {
		return nil, fmt.Errorf("Error recording service proof for application %s: %w", relay.Meta.ApplicationAddress, err)
	}

	if err := localCtx.Release(); err != nil {
		s.logger.Warn().Err(err).Msg("failed to release local context")
	}

	return response, nil
}

// isRelayVolumeApplicable returns:
//  1. The signed digest of a relay/response pair,
//  2. Whether there was a collision for the specific chain (i.e. should the service proof be stored for claiming later)
func (s *servicer) isRelayVolumeApplicable(relay *coreTypes.Relay, response *coreTypes.RelayResponse) (digest, serializedRelayRes []byte, collides bool, err error) {
	relayReqResBytes, err := codec.GetCodec().Marshal(&coreTypes.RelayReqRes{Relay: relay, Response: response})
	if err != nil {
		return nil, nil, false, fmt.Errorf("Error marshalling relay and/or response: %w", err)
	}

	relayDigest := crypto.SHA3Hash(relayReqResBytes)

	signedDigest := s.sign(relayDigest)
	response.ServicerSignature = hex.EncodeToString(signedDigest)
	collision, err := s.isRelayVolumeApplicableOnChain(relayDigest, relay.Meta.RelayChain.Id)
	if err != nil {
		return nil, nil, false, fmt.Errorf("Error checking collision for chain %s: %w", relay.Meta.RelayChain.Id, err)
	}

	return signedDigest, relayReqResBytes, collision, nil
}

// INCOMPLETE: implement this
func (s *servicer) sign(bz []byte) []byte {
	return bz
}

// INCOMPLETE: implement this
func (s *servicer) isRelayVolumeApplicableOnChain(digest []byte, relayChainId string) (bool, error) {
	return false, nil
}

// executeRelay performs the passed relay using the correct method depending on the relay payload type.
func (s *servicer) executeRelay(relay *coreTypes.Relay) (*coreTypes.RelayResponse, error) {
	switch payload := relay.RelayPayload.(type) {
	case *coreTypes.Relay_JsonRpcPayload:
		return s.executeHTTPRelay(relay.Meta, payload.JsonRpcPayload)
	case *coreTypes.Relay_RestPayload:
		return nil, fmt.Errorf("Error executing relay on application %s: REST not supported", relay.Meta.ApplicationAddress)
	default:
		return nil, fmt.Errorf("Error exeucting relay on application %s: Unsupported type on payload %s", relay.Meta.ApplicationAddress, payload)
	}
}

// validateRelayMeta ensures the relay metadata is valid for being handled by the servicer
// REFACTOR: move the meta-specific validation to a Validator method on RelayMeta struct
func (s *servicer) validateRelayMeta(meta *coreTypes.RelayMeta, currentHeight int64) error {
	if meta == nil {
		return fmt.Errorf("empty relay metadata")
	}

	if meta.RelayChain == nil {
		return fmt.Errorf("relay chain unspecified")
	}

	if err := s.validateRelayChainSupport(meta.RelayChain, currentHeight); err != nil {
		return fmt.Errorf("validation of support for relay chain %s failed: %w", meta.RelayChain.Id, err)
	}

	return nil
}

func (s *servicer) validateRelayChainSupport(relayChain *coreTypes.Identifiable, currentHeight int64) error {
	if _, ok := s.config.Services[relayChain.Id]; !ok {
		return fmt.Errorf("service %s not supported by servicer %s configuration", relayChain.Id, s.config.Address)
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

// validateApplication makes sure the application has not received more relays than allocated in the current session.
func (s *servicer) validateApplication(session *coreTypes.Session) error {
	// IMPROVE: use a function to get current height from the current session
	servicerAppSessionTokens, err := s.calculateServicerAppSessionTokens(session)
	if err != nil {
		return fmt.Errorf("Error calculating servicer tokens for application: %w", err)
	}

	localCtx, err := s.GetBus().GetPersistenceModule().NewLocalContext()
	if err != nil {
		return fmt.Errorf("Error getting local persistence context: application %s session number %d: %w", session.Application.PublicKey, session.SessionNumber, err)
	}

	usedAppSessionTokens, err := localCtx.GetSessionTokensUsed(session)
	if err != nil {
		return fmt.Errorf("Error getting servicer token usage: application %s session number %d: %w", session.Application.PublicKey, session.SessionNumber, err)
	}

	if usedAppSessionTokens == nil || usedAppSessionTokens.Cmp(servicerAppSessionTokens) < 0 {
		return nil
	}

	return fmt.Errorf("application %s has exceeded its allocated relays %s for session %d",
		session.Application.PublicKey,
		servicerAppSessionTokens,
		session.SessionNumber)
}

func (s *servicer) cachedAppTokens(session *coreTypes.Session) *sessionTokens {
	s.rwlock.RLock()
	defer s.rwlock.RUnlock()

	return s.totalTokens[session.Application.PublicKey]
}

// calculateServicerAppSessionTokens return the number of tokens the servicer has remaining for the Application in the session provided.
//
//	If nothing is cached, the maximum number of session tokens is computed.
func (s *servicer) calculateServicerAppSessionTokens(session *coreTypes.Session) (*big.Int, error) {
	tokens := s.cachedAppTokens(session)
	if tokens != nil && tokens.originalCountAvailable != nil && tokens.sessionNumber == session.SessionNumber {
		return big.NewInt(1).Set(tokens.originalCountAvailable), nil
	}

	// Calculate this servicer's limit for the application in the current session.
	//	This is distributed rate limiting (DRL): no need to know how many requests have
	//		been performed for this application by other servicers. Instead, simply enforce
	//		this servicer's share of the application's tokens for this session.
	appSessionTokens, err := s.calculateAppSessionTokens(session)
	if err != nil {
		return nil, fmt.Errorf("Error calculating application %s total tokens for session %d: %w", session.Application.PublicKey, session.SessionNumber, err)
	}

	// type conversion from big.Int to big.Float
	appTokens := big.NewFloat(1).SetInt(appSessionTokens)
	servicerTokens := appTokens.Quo(appTokens, big.NewFloat(float64(len(session.Servicers))))

	// This multiplication is performed to minimize the chance of under-utilization of application's tokens,
	//	while removing the overhead of communication between servicers which would be necessary otherwise.
	// see https://arxiv.org/abs/2305.10672 for details on application and servicer distributed rate-limiting
	adjustedTokens := servicerTokens.Mul(servicerTokens, big.NewFloat(1+s.config.RelayMiningVolumeAccuracy))
	roundedTokens, _ := adjustedTokens.Int(big.NewInt(1))

	s.setAppSessionTokens(session, &sessionTokens{session.SessionNumber, roundedTokens})
	return roundedTokens, nil
}

func (s *servicer) setAppSessionTokens(session *coreTypes.Session, tokens *sessionTokens) {
	s.rwlock.Lock()
	defer s.rwlock.Unlock()

	if len(s.totalTokens) == 0 {
		s.totalTokens = make(map[string]*sessionTokens)
	}

	s.totalTokens[session.Application.PublicKey] = tokens
}

// validateServicer makes sure the servicer is A) active in the current session, and B) has not served more than its allocated relays for the session
func (s *servicer) validateServicer(meta *coreTypes.RelayMeta, session *coreTypes.Session) error {
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

	return nil
}

// getSession returns a session for the current height and the passed relay
func (s *servicer) getSession(relay *coreTypes.Relay) (*coreTypes.Session, error) {
	height := s.GetBus().GetConsensusModule().CurrentHeight()
	session, err := s.GetBus().GetUtilityModule().GetSession(relay.Meta.ApplicationAddress, int64(height), relay.Meta.RelayChain.Id, relay.Meta.GeoZone.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get a session for height %d for relay meta %s: %w", height, relay.Meta, err)
	}

	return session, nil
}

// admitRelay decides whether the relay should be served
func (s *servicer) admitRelay(relay *coreTypes.Relay) error {
	// TODO: utility module should initialize the servicer (if this module instance is a servicer)
	const errPrefix = "Error admitting relay"

	if relay == nil {
		return fmt.Errorf("%s: relay is nil", errPrefix)
	}

	height := s.GetBus().GetConsensusModule().CurrentHeight()
	if err := s.validateRelayMeta(relay.Meta, int64(height)); err != nil {
		return fmt.Errorf("%s: %w", err.Error(), errValidateRelayMeta)
	}

	// TODO: update the CLI to include ApplicationAddress(or Application Public Key) in the RelayMeta
	session, err := s.getSession(relay)
	if err != nil {
		return err
	}

	// TODO: (REFACTOR) use a loop to run all validators: would also remove the need for passing the session around
	if err := validateRelayBlockHeight(relay.Meta, session); err != nil {
		return fmt.Errorf("%s: %w", err.Error(), errValidateBlockHeight)
	}

	if err := s.validateServicer(relay.Meta, session); err != nil {
		return fmt.Errorf("%s: %s: %w", errPrefix, err.Error(), errValidateServicer)
	}

	if err := s.validateApplication(session); err != nil {
		return fmt.Errorf("%s: %s: %w", errPrefix, err.Error(), errValidateApplication)
	}

	return nil
}

// DISCUSS: do we need to export this functionality as part of the utility module?
// calculateAppSessionTokens determines the number of "session tokens" an application gets at the beginning
// of every session. Each servicer will serve a maximum of ~(Session Tokens / Number of Servicers in the Session) relays for the application
func (s *servicer) calculateAppSessionTokens(session *coreTypes.Session) (*big.Int, error) {
	appStake, err := utils.StringToBigInt(session.Application.StakedAmount)
	if err != nil {
		return nil, fmt.Errorf("Error processing application's staked amount %s: %w", session.Application.StakedAmount, coreTypes.ErrStringToBigInt(err))
	}

	// DOCUMENT: find the right document to explain the following:
	//	We assume that the value of certain parameters only changes/takes effect at the start of a session.
	//	In this specific case, the `AppSessionTokensMultiplierParamName` parameter is retrieved for the height that
	//		matches the beginning of the session.
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(session.SessionHeight)
	if err != nil {
		return nil, fmt.Errorf("error getting persistence context at height %d: %w", session.SessionHeight, err)
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	// DISCUSS: using an interface for returning each defined parameter seems less error-prone: also could return e.g. int64 in this case to remove the type cast
	appStakeTokensMultiplier, err := readCtx.GetIntParam(typesUtil.AppSessionTokensMultiplierParamName, session.SessionHeight)
	if err != nil {
		return nil, fmt.Errorf("error reading parameter %s at height %d from persistence: %w", typesUtil.AppSessionTokensMultiplierParamName, session.SessionHeight, err)
	}

	return appStake.Mul(appStake, big.NewInt(int64(appStakeTokensMultiplier))), nil
}

// executeHTTPRequest performs the HTTP request that sends the relay to the chain's URL.
func (s *servicer) executeHTTPRelay(meta *coreTypes.RelayMeta, payload *coreTypes.JsonRpcPayload) (*coreTypes.RelayResponse, error) {
	if meta == nil || meta.RelayChain == nil || meta.RelayChain.Id == "" {
		return nil, fmt.Errorf("Relay for application %s does not specify relay chain", meta.ApplicationAddress)
	}

	serviceConfig, ok := s.config.Services[meta.RelayChain.Id]
	if !ok {
		return nil, fmt.Errorf("Chain %s not found in servicer configuration: %w", meta.RelayChain.Id, errValidateRelayMeta)
	}

	chainUrl, err := url.Parse(serviceConfig.Url)
	if err != nil {
		return nil, fmt.Errorf("Error parsing chain URL %s: %w", serviceConfig.Url, err)
	}
	targetUrl := chainUrl.JoinPath(payload.HttpPath)

	req, err := http.NewRequest(payload.Method, targetUrl.String(), bytes.NewBuffer([]byte(payload.Data)))
	if err != nil {
		return nil, err
	}
	if serviceConfig.BasicAuth != nil && serviceConfig.BasicAuth.UserName != "" {
		req.SetBasicAuth(serviceConfig.BasicAuth.UserName, serviceConfig.BasicAuth.Password)
	}

	// DISCUSS: do we need a default user-agent for HTTP requests?
	for k, v := range payload.Headers {
		req.Header.Set(k, v)
	}
	if len(payload.Headers) == 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	// DISCUSS: we need to optimize usage of HTTP client, e.g. for connection reuse, considering the expected volume of relays
	resp, err := (&http.Client{Timeout: time.Duration(serviceConfig.TimeoutMsec) * time.Millisecond}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error performing the HTTP request for relay: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %w", err)
	}

	return &coreTypes.RelayResponse{Payload: string(body)}, nil
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
