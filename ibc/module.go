package ibc

import (
	"encoding/hex"
	"fmt"
	"sync"

	ibcTypes "github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.IBCModule = &ibcModule{}

type ibcModule struct {
	base_modules.IntegratableModule

	m sync.Mutex

	cfg    *configs.IBCConfig
	logger *modules.Logger

	// Only a single host is allowed at a time
	host *host
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(ibcModule).Create(bus, options...)
}

func (m *ibcModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	*m = ibcModule{
		cfg:    bus.GetRuntimeMgr().GetConfig().IBC,
		logger: logger.Global.CreateLoggerForModule(modules.IBCModuleName),
	}
	m.logger.Info().Msg("🪐 creating IBC module 🪐")

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	// Only validators can be an IBC host due to the need for reliability
	isValidator := false
	if _, err := m.GetBus().GetUtilityModule().GetValidatorModule(); err == nil {
		isValidator = true
	}
	if isValidator && m.cfg.Enabled {
		m.logger.Info().Msg("🛰️ creating IBC host 🛰️")
		if err := m.newHost(); err != nil {
			m.logger.Error().Err(err).Msg("❌ failed to create IBC host")
			return nil, err
		}
	}

	return m, nil
}

func (m *ibcModule) Start() error {
	if !m.cfg.Enabled {
		m.logger.Info().Msg("🚫 IBC module disabled 🚫")
		return nil
	}
	m.logger.Info().Msg("🪐 starting IBC module 🪐")
	// TODO: start the host logic
	return nil
}

func (m *ibcModule) Stop() error {
	return nil
}

func (m *ibcModule) GetHost() modules.IBCHost {
	return m.host
}

func (m *ibcModule) GetModuleName() string {
	return modules.IBCModuleName
}

// HandleMessage accepts a generic IBC message and routes it to the specific handler
func (m *ibcModule) HandleMessage(message *anypb.Any) error {
	m.m.Lock()
	defer m.m.Unlock()

	// Check the message is actually a valid IBC message
	msg, err := codec.GetCodec().FromAny(message)
	if err != nil {
		return err
	}
	ibcMessage, ok := msg.(*ibcTypes.IBCMessage)
	if !ok {
		return fmt.Errorf("failed to cast message to IBCMessage")
	}
	if err := ibcMessage.ValidateBasic(); err != nil {
		return err
	}

	// Convert IBC message to a utility Transaction
	tx, err := ConvertIBCMessageToTx(ibcMessage)
	if err != nil {
		return err
	}

	// Sign the transaction
	pkBz, err := hex.DecodeString(m.cfg.PrivateKey)
	if err != nil {
		return err
	}
	pk, err := crypto.NewPrivateKeyFromBytes(pkBz)
	if err != nil {
		return err
	}
	signableBz, err := tx.SignableBytes()
	if err != nil {
		return err
	}
	signature, err := pk.Sign(signableBz)
	if err != nil {
		return err
	}
	tx.Signature = &coreTypes.Signature{
		Signature: signature,
		PublicKey: pk.PublicKey().Bytes(),
	}

	// Marshall the Transaction and send it to the utility module
	txBz, err := codec.GetCodec().Marshal(tx)
	if err != nil {
		return err
	}
	if err := m.GetBus().GetUtilityModule().HandleTransaction(txBz); err != nil {
		return err
	}
	m.logger.Info().Str("message_type", "IBCMessage").Msg("Successfully added a new message to the mempool!")
	return nil
}

// newHost returns a new IBC host instance if one is not already created
func (m *ibcModule) newHost() error {
	if m.host != nil {
		return coreTypes.ErrHostAlreadyExists()
	}
	host := &host{
		logger: m.logger,
	}
	m.host = host
	return nil
}
