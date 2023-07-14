package client

import (
	"fmt"

	"github.com/pokt-network/pocket/ibc/path"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_                  modules.ClientManager = &clientManager{}
	allowedClientTypes                       = make(map[string]struct{}, 0)
)

func init() {
	allowedClientTypes["08-wasm"] = struct{}{}
}

type clientManager struct {
	base_modules.IntegrableModule

	logger *modules.Logger
}

func Create(bus modules.Bus, options ...modules.ClientManagerOption) (modules.ClientManager, error) {
	return new(clientManager).Create(bus, options...)
}

// WithLogger sets the logger for the clientManager
func WithLogger(logger *modules.Logger) modules.ClientManagerOption {
	return func(m modules.ClientManager) {
		if mod, ok := m.(*clientManager); ok {
			mod.logger = logger
		}
	}
}

func (*clientManager) Create(bus modules.Bus, options ...modules.ClientManagerOption) (modules.ClientManager, error) {
	c := &clientManager{}

	for _, option := range options {
		option(c)
	}

	c.logger.Info().Msg("👨 Creating Client Manager 👨")

	bus.RegisterModule(c)

	return c, nil
}

func (c *clientManager) GetModuleName() string { return modules.ClientManagerModuleName }

// CreateClient creates a new client with the given client state and initial
// consensus state and initialises it with a unique identifier in the IBC client
// store and emits an event to the Event Logger
func (c *clientManager) CreateClient(
	clientState modules.ClientState, consensusState modules.ConsensusState,
) (string, error) {
	// Check if the client type is allowed
	if !isAllowedClientType(clientState.ClientType()) {
		return "", fmt.Errorf("client type %s is not supported", clientState.ClientType())
	}

	// Generate a unique identifier for the client
	identifier := path.GenerateClientIdentifier()

	// Retrieve the client store
	clientStore, err := c.GetBus().GetIBCHost().GetProvableStore(path.KeyClientStorePrefix)
	if err != nil {
		return "", err
	}

	// Initialise the client with the clientState provided
	if err := clientState.Initialise(clientStore, consensusState); err != nil {
		c.logger.Error().Err(err).Str("identifier", identifier).Msg("failed to initialize client")
		return "", err
	}

	c.logger.Info().Str("identifier", identifier).Str("height", clientState.GetLatestHeight().String()).Msg("client created at height")

	// Emit the create client event to the event logger
	if err := c.emitCreateClientEvent(identifier, clientState); err != nil {
		c.logger.Error().Err(err).Str("identifier", identifier).Msg("failed to emit client created event")
		return "", err
	}

	return identifier, nil
}

// UpdateClient updates an existing client with the given identifer using the
// ClientMessage provided
func (c *clientManager) UpdateClient(
	identifier string, clientMessage modules.ClientMessage,
) error {
	return nil
}

// QueryConsensusState returns the ConsensusState at the given height for the
// stored client with the given identifier
func (c *clientManager) QueryConsensusState(
	identifier string, height modules.Height,
) (modules.ConsensusState, error) {
	return nil, nil
}

// QueryClientState returns the ClientState for the stored client with the given identifier
func (c *clientManager) QueryClientState(identifier string) (modules.ClientState, error) {
	return nil, nil
}

// SubmitMisbehaviour submits evidence for a misbehaviour to the client, possibly
// invalidating previously valid state roots and thus preventing future updates
func (c *clientManager) SubmitMisbehaviour(
	identifier string, clientMessage modules.ClientMessage,
) error {
	return nil
}

func isAllowedClientType(clientType string) bool {
	if _, ok := allowedClientTypes[clientType]; ok {
		return true
	}
	return false
}
