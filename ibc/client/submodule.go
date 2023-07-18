package client

import (
	"fmt"

	"github.com/pokt-network/pocket/ibc/client/types"
	"github.com/pokt-network/pocket/ibc/path"
	core_types "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_                  modules.ClientManager = &clientManager{}
	allowedClientTypes                       = make(map[string]struct{}, 0)
)

func init() {
	allowedClientTypes[types.WasmClientType] = struct{}{}
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

	// Retrieve the client store prefixed with the client identifier
	prefixed := path.ApplyPrefix(core_types.CommitmentPrefix(path.KeyClientStorePrefix), identifier)
	clientStore, err := c.GetBus().GetIBCHost().GetProvableStore(string(prefixed))
	if err != nil {
		return "", err
	}

	// Initialise the client with the clientState provided
	if err := clientState.Initialise(clientStore, consensusState); err != nil {
		c.logger.Error().Err(err).Str("identifier", identifier).
			Msg("failed to initialize client")
		return "", err
	}

	c.logger.Info().Str("identifier", identifier).
		Str("height", clientState.GetLatestHeight().ToString()).
		Msg("client created at height")

	// Emit the create client event to the event logger
	if err := c.emitCreateClientEvent(identifier, clientState); err != nil {
		c.logger.Error().Err(err).Str("identifier", identifier).
			Msg("failed to emit client created event")
		return "", err
	}

	return identifier, nil
}

// UpdateClient updates an existing client with the given identifier using the
// ClientMessage provided
func (c *clientManager) UpdateClient(
	identifier string, clientMessage modules.ClientMessage,
) error {
	// Get the client state
	clientState, err := c.GetClientState(identifier)
	if err != nil {
		return err
	}

	// Retrieve the client store prefixed with the client identifier
	prefixed := path.ApplyPrefix(core_types.CommitmentPrefix(path.KeyClientStorePrefix), identifier)
	clientStore, err := c.GetBus().GetIBCHost().GetProvableStore(string(prefixed))
	if err != nil {
		return err
	}

	// Check the state is active
	if clientState.Status(clientStore) != modules.ActiveStatus {
		return core_types.ErrIBCClientNotActive()
	}

	// Verify the client message
	if err := clientState.VerifyClientMessage(clientStore, clientMessage); err != nil {
		return err
	}

	// Check for misbehaviour on the source chain
	misbehaved := clientState.CheckForMisbehaviour(clientStore, clientMessage)
	if misbehaved {
		if err := clientState.UpdateStateOnMisbehaviour(clientStore, clientMessage); err != nil {
			c.logger.Error().Err(err).Str("identifier", identifier).
				Msg("failed to freeze client for misbehaviour")
			return err
		}
		c.logger.Info().Str("identifier", identifier).
			Msg("client frozen for misbehaviour")

		// emit the submit misbehaviour event to the event logger
		if err := c.emitSubmitMisbehaviourEvent(identifier, clientState); err != nil {
			c.logger.Error().Err(err).Str("identifier", identifier).
				Msg("failed to emit client submit misbehaviour event")
			return err
		}
		return nil
	}

	// Update the client
	consensusHeight, err := clientState.UpdateState(clientStore, clientMessage)
	if err != nil {
		c.logger.Error().Err(err).Str("identifier", identifier).
			Str("height", consensusHeight.ToString()).
			Msg("failed to update client state")
		return err
	}
	c.logger.Info().Str("identifier", identifier).
		Str("height", consensusHeight.ToString()).
		Msg("client state updated")

	// emit the update client event to the event logger
	if err := c.emitUpdateClientEvent(identifier, clientState.ClientType(), consensusHeight, clientMessage); err != nil {
		c.logger.Error().Err(err).Str("identifier", identifier).
			Msg("failed to emit client update event")
		return err
	}

	return nil
}

// UpgradeClient upgrades an existing client with the given identifier using the
// ClientState and ConsentusState provided. It can only do so if the new client
// was committed to by the old client at the specified upgrade height
func (c *clientManager) UpgradeClient(
	identifier string,
	upgradedClient modules.ClientState, upgradedConsState modules.ConsensusState,
	proofUpgradeClient, proofUpgradeConsState []byte,
) error {
	// Get the client state
	clientState, err := c.GetClientState(identifier)
	if err != nil {
		return err
	}

	// Retrieve the client store prefixed with the client identifier
	prefixed := path.ApplyPrefix(core_types.CommitmentPrefix(path.KeyClientStorePrefix), identifier)
	clientStore, err := c.GetBus().GetIBCHost().GetProvableStore(string(prefixed))
	if err != nil {
		return err
	}

	// Check the state is active
	if clientState.Status(clientStore) != modules.ActiveStatus {
		return core_types.ErrIBCClientNotActive()
	}

	// Verify the upgrade
	if err := clientState.VerifyUpgradeAndUpdateState(
		clientStore,
		upgradedClient, upgradedConsState,
		proofUpgradeClient, proofUpgradeConsState,
	); err != nil {
		c.logger.Error().Err(err).Str("identifier", identifier).
			Msg("failed to verify upgrade")
		return err
	}

	c.logger.Info().Str("identifier", identifier).
		Str("height", upgradedClient.GetLatestHeight().ToString()).
		Msg("client upgraded")

	// emit the upgrade client event to the event logger
	if err := c.emitUpgradeClientEvent(identifier, upgradedClient); err != nil {
		c.logger.Error().Err(err).Str("identifier", identifier).
			Msg("failed to emit client upgrade event")
		return err
	}

	return nil
}

func isAllowedClientType(clientType string) bool {
	if _, ok := allowedClientTypes[clientType]; ok {
		return true
	}
	return false
}
