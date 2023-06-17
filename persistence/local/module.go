package local

import (
	"math/big"

	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

const (
	LocalModuleName = "local"
)

var _ modules.PersistenceLocalContext = &persistenceLocalContext{}

type persistenceLocalContext struct {
	base_modules.IntegratableModule

	logger       *modules.Logger
	databasePath string
}

func WithLocalContextConfig(databasePath string) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		if plc, ok := m.(*persistenceLocalContext); ok {
			plc.databasePath = databasePath
		}
	}
}

func CreateLocalContext(bus modules.Bus, options ...modules.ModuleOption) (modules.PersistenceLocalContext, error) {
	m, err := new(persistenceLocalContext).Create(bus, options...)
	if err != nil {
		return nil, err
	}
	return m.(modules.PersistenceLocalContext), nil
}

func (*persistenceLocalContext) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &persistenceLocalContext{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	return m, nil
}

func (m *persistenceLocalContext) GetModuleName() string {
	return LocalModuleName
}

// INCOMPLETE(#826): implement this
func (m *persistenceLocalContext) Start() error {
	return nil
}

// INCOMPLETE(#826): implement this
func (m *persistenceLocalContext) Stop() error {
	return nil
}

// INCOMPLETE(#826): implement this
// OPTIMIZE: both the relay and the response can be large structures: we may need to truncate the stored values
// StoreServicedRelay implements the PersistenceLocalContext interface
func (local *persistenceLocalContext) StoreServicedRelay(session *coreTypes.Session, relayDigest, relayReqResBytes []byte) error {
	return nil
}

// INCOMPLETE(#826): implement this
// GetSessionTokensUsed implements the PersistenceLocalContext interface
func (local *persistenceLocalContext) GetSessionTokensUsed(*coreTypes.Session) (*big.Int, error) {
	return nil, nil
}
