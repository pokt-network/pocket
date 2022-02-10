package utility

import (
	"log"

	"pocket/consensus/pkg/shared/context"
	"pocket/consensus/pkg/shared/modules"
	"pocket/consensus/pkg/types/typespb"
)

type utilityModule struct {
	*modules.BasePocketModule
	modules.UtilityModule
}

func Create(ctx *context.PocketContext, base *modules.BasePocketModule) (m modules.UtilityModule, err error) {
	log.Println("Creating utility module")
	m = &utilityModule{
		BasePocketModule: base,
	}
	return m, nil
}

func (m *utilityModule) Start(ctx *context.PocketContext) error {
	log.Println("Starting utility module")
	return nil
}

func (m *utilityModule) Stop(ctx *context.PocketContext) error {
	log.Println("Stopping utility module")
	return nil
}

func (m *utilityModule) GetPocketBusMod() modules.PocketBusModule {
	return m.BasePocketModule.GetPocketBusMod()
}

func (m *utilityModule) SetPocketBusMod(bus modules.PocketBusModule) {
	m.BasePocketModule.SetPocketBusMod(bus)
}

func (m *utilityModule) HandleTransaction(*context.PocketContext, *typespb.Transaction) error {
	log.Println("[TODO] Utility HandleTransaction not implemented yet...")
	return nil
}

func (m *utilityModule) HandleEvidence(*context.PocketContext, *typespb.Evidence) error {
	log.Println("[TODO] Utility HandleEvidence not implemented yet...")
	return nil
}

func (m *utilityModule) ReapMempool(*context.PocketContext) ([]*typespb.Transaction, error) {
	log.Println("[TODO] Utility ReapMempool not implemented yet...")
	return make([]*typespb.Transaction, 0), nil
}

func (m *utilityModule) BeginBlock(*context.PocketContext) error {
	log.Println("[TODO] Utility BeginBlock not implemented yet...")
	return nil
}

func (m *utilityModule) DeliverTx(*context.PocketContext, *typespb.Transaction) error {
	log.Println("[TODO] Utility DeliverTx not implemented yet...")
	return nil
}

func (m *utilityModule) EndBlock(*context.PocketContext) error {
	log.Println("[TODO] Utility EndBlock not implemented yet...")
	return nil
}
