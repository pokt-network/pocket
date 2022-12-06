package utility

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.UtilityModule = &utilityModule{}
var _ modules.UtilityConfig = &types.UtilityConfig{}
var _ modules.Module = &utilityModule{}

type utilityModule struct {
	bus    modules.Bus
	config modules.UtilityConfig

	Mempool types.Mempool
}

const (
	utilityModuleName = "utility"

	TransactionGossipContentType = "utility.TransactionGossip"
)

func Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	return new(utilityModule).Create(runtime)
}

func (*utilityModule) Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	var m *utilityModule

	cfg := runtime.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	utilityCfg := cfg.GetUtilityConfig()

	return &utilityModule{
		config:  utilityCfg,
		Mempool: types.NewMempool(utilityCfg.GetMaxMempoolTransactionBytes(), utilityCfg.GetMaxMempoolTransactions()),
	}, nil
}

func (u *utilityModule) Start() error {
	return nil
}

func (u *utilityModule) Stop() error {
	return nil
}

func (u *utilityModule) GetModuleName() string {
	return utilityModuleName
}

func (u *utilityModule) SetBus(bus modules.Bus) {
	u.bus = bus
}

func (u *utilityModule) GetBus() modules.Bus {
	if u.bus == nil {
		log.Fatalf("Bus is not initialized")
	}
	return u.bus
}

func (*utilityModule) ValidateConfig(cfg modules.Config) error {
	// TODO (#334): implement this
	return nil
}

func (u *utilityModule) HandleMessage(message *anypb.Any) error {
	switch message.MessageName() {
	case TransactionGossipContentType:
		msg, err := codec.GetCodec().FromAny(message)
		if err != nil {
			return err
		}
		transactionGossipMsg, ok := msg.(*types.TransactionGossip)
		if !ok {
			return fmt.Errorf("failed to cast message to UtilityMessage")
		}
		if err := u.CheckTransaction(transactionGossipMsg.Tx); err != nil {
			return err
		}
		log.Println("MEMPOOOL: Successfully added a new message to the mempool!")
	default:
		return types.ErrUnknownMessageType(message.MessageName())
	}
	return nil
}
