package p2p

import (
	"log"
	"pocket/shared/config"
	"pocket/shared/modules"
	"pocket/shared/types"
)

type P2PModule struct {
	modules.NetworkModule
	pocketBusMod modules.BusModule

	p2pConfig *config.P2PConfig

	gater *gater
}

func Create(config *config.Config) (modules.NetworkModule, error) {
	return &P2PModule{
		p2pConfig: config.P2P,
	}, nil
}

func (p *P2PModule) Start() error {
	gater := NewGater()
	gater.Config(p.p2pConfig.Protocol, p.p2pConfig.Address, p.p2pConfig.ExternalIp, p.p2pConfig.Peers)
	gater.Init()

	go gater.Listen()

	<-gater.Ready()

	p.gater = gater

	return nil
}

func (p *P2PModule) Stop() error {
	return nil
}

func (m *P2PModule) SetPocketBusMod(pocketBus modules.BusModule) {
	m.pocketBusMod = pocketBus
}

func (m *P2PModule) GetBus() modules.BusModule {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (m *P2PModule) BroadcastMessage(msg *types.NetworkMessage) error {
	return m.gater.BroadcastTempWrapper(msg)
}
