package p2p

import (
	"errors"
	"fmt"
	"log"
	"pocket/consensus/pkg/config"
	"pocket/shared/context"
	"pocket/shared/messages"
	"pocket/shared/modules"
)

type P2PModule struct {
	modules.NetworkModule
	pocketBusMod modules.PocketBusModule

	p2pConfig *config.P2PConfig

	gater *gater
}

var (
	ErrNotCreated = errors.New("Module error: P2P Module not created. Trying to start the p2p module before calling create.")
)

func Create(config *config.Config) (modules.NetworkModule, error) {
	g := NewGater()
	cfg := config.P2P

	return &P2PModule{
		p2pConfig: cfg,
		gater:     g,
	}, nil
}

func (p *P2PModule) Start(ctx *context.PocketContext) error {
	if p.gater == nil {
		return ErrNotCreated
	}

	p.gater.Config(p.p2pConfig.Protocol, p.p2pConfig.Address, p.p2pConfig.ExternalIp, p.p2pConfig.Peers)
	fmt.Println("p.gater.list", p.gater.peerlist)
	p.gater.Init()

	go p.gater.Listen()

	<-p.gater.Ready()

	return nil
}

func (p *P2PModule) Stop(*context.PocketContext) error {
	p.gater.Close()

	<-p.gater.closed
	<-p.gater.done

	return nil
}

func (p *P2PModule) SetPocketBusMod(pocketBus modules.PocketBusModule) {
	p.pocketBusMod = pocketBus
}

func (p *P2PModule) GetPocketBusMod() modules.PocketBusModule {
	if p.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return p.pocketBusMod
}

func (p *P2PModule) BroadcastMessage(msg *messages.NetworkMessage) error {
	return p.gater.Broadcast(msg, true)
}

func (p *P2PModule) Send(addr string, msg *messages.NetworkMessage) error {
	encoded, err := p.gater.c.encode(msg)
	if err != nil {
		return err
	}

	return p.gater.Send(addr, encoded, true) // true: meaning that this message is already encoded
}

func (p *P2PModule) AckSend(addr string, msg *messages.NetworkMessage) (bool, error) {
	encoded, err := p.gater.c.encode(msg)
	if err != nil {
		return false, err
	}

	response, err := p.gater.Request(addr, encoded, true) // true: meaning that this message is already encoded
	if err != nil {
		return false, err
	}

	ack, err := p.gater.c.decode(response)
	if err != nil {
		return true, err // TODO: notice it's true
	}

	ackmsg := ack.(*messages.NetworkMessage)

	if ackmsg.Nonce == msg.Nonce {
		return true, nil
	}

	return false, nil
}
