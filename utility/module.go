package utility

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.UtilityModule = &utilityModule{}
var _ modules.Module = &utilityModule{}

type utilityModule struct {
	bus    modules.Bus
	config *configs.UtilityConfig

	Mempool types.Mempool
}

const (
	TransactionGossipMessageContentType = "utility.TransactionGossipMessage"
)

func Create(bus modules.Bus) (modules.Module, error) {
	return new(utilityModule).Create(bus)
}

func (*utilityModule) Create(bus modules.Bus) (modules.Module, error) {
	m := &utilityModule{}
	if err := bus.RegisterModule(m); err != nil {
		return nil, err
	}

	runtimeMgr := bus.GetRuntimeMgr()

	cfg := runtimeMgr.GetConfig()
	utilityCfg := cfg.Utility

	m.config = utilityCfg
	m.Mempool = types.NewMempool(utilityCfg.MaxMempoolTransactionBytes, utilityCfg.MaxMempoolTransactions)

	return m, nil
}

func (u *utilityModule) Start() error {
	return nil
}

func (u *utilityModule) Stop() error {
	return nil
}

func (u *utilityModule) GetModuleName() string {
	return modules.UtilityModuleName
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

func (u *utilityModule) HandleMessage(message *anypb.Any) error {
	switch message.MessageName() {
	case TransactionGossipMessageContentType:
		msg, err := codec.GetCodec().FromAny(message)
		if err != nil {
			return err
		}
		transactionGossipMsg, ok := msg.(*types.TransactionGossipMessage)
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
