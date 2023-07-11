package client

import (
	"github.com/pokt-network/pocket/ibc/path"
	core_types "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.ClientManager = &clientManager{}

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
	identifier := path.GenerateClientIdentifier()
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
	identifier string, height *core_types.Height,
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
