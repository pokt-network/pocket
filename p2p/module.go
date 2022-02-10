package p2p

import (
	"log"
	"pocket/consensus/pkg/config"
	"pocket/shared/context"
	"pocket/shared/modules"
)

type P2PModule struct {
	modules.NetworkModule

	pocketBusMod modules.PocketBusModule
	gater *gater

	p2pConfig *config.P2PConfig
}

func Create(config *config.Config) (modules.NetworkModule, error) {
	return &P2PModule{
		p2pConfig: config.P2P,
	}, nil
}


func(p *P2PModule) Start(ctx *context.PocketContext) error {
	p.gater = NewGater()
	p.gater.Config(p.p2pConfig.Protocol, p.p2pConfig.Address, p.p2pConfig.ExternalIp, p.p2pConfig.Peers)
	p.gater.Init()

	go p.gater.Listen()

	<- p.gater.Ready()

	return nil
}

func(p *P2PModule) Stop(*context.PocketContext) error {
	return nil
}

func (m *P2PModule) SetPocketBusMod(pocketBus modules.PocketBusModule) {
	m.pocketBusMod = pocketBus
}

func (m *P2PModule) GetPocketBusMod() modules.PocketBusModule {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}