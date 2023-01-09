package debug

import (
	"log"

	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	"github.com/pokt-network/pocket/p2p/transport"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const (
	// ANY_HEIGHT is a special height that will be used to indicate that the actors are valid for all heights (including future heights)
	ANY_HEIGHT = -1
)

var _ addrbook_provider.AddrBookProvider = &debugAddrBookProvider{}

type debugAddrBookProvider struct {
	p2pCfg         *configs.P2PConfig
	actorsByHeight map[int64][]*coreTypes.Actor
	connFactory    typesP2P.ConnectionFactory
}

func NewDebugAddrBookProvider(p2pCfg *configs.P2PConfig, options ...func(*debugAddrBookProvider)) *debugAddrBookProvider {
	dabp := &debugAddrBookProvider{
		p2pCfg:      p2pCfg,
		connFactory: transport.CreateDialer, // default connection factory, overridable with WithConnectionFactory()
	}

	for _, o := range options {
		o(dabp)
	}

	return dabp
}

func WithActorsByHeight(actorsByHeight map[int64][]*coreTypes.Actor) func(*debugAddrBookProvider) {
	return func(pabp *debugAddrBookProvider) {
		pabp.actorsByHeight = actorsByHeight
	}
}

func (dabp *debugAddrBookProvider) getActorsByHeight(height uint64) []*coreTypes.Actor {
	if stakedActors, ok := dabp.actorsByHeight[ANY_HEIGHT]; ok {
		log.Println("[DEBUG] Ignoring height param in debugAddrBookProvider")
		return stakedActors
	}

	if stakedActors, ok := dabp.actorsByHeight[int64(height)]; ok {
		return stakedActors
	}

	log.Fatalf("No actors found for height %d. Please make sure you configured the provider via WithActorsByHeight", height)
	return nil
}

func (dabp *debugAddrBookProvider) GetStakedAddrBookAtHeight(height uint64) (typesP2P.AddrBook, error) {
	stakedActors := dabp.getActorsByHeight(height)
	return addrbook_provider.ActorsToAddrBook(dabp, stakedActors)
}

func (dabp *debugAddrBookProvider) GetConnFactory() typesP2P.ConnectionFactory {
	return dabp.connFactory
}

func (dabp *debugAddrBookProvider) GetP2PConfig() *configs.P2PConfig {
	return dabp.p2pCfg
}

func (dabp *debugAddrBookProvider) SetConnectionFactory(connFactory typesP2P.ConnectionFactory) {
	dabp.connFactory = connFactory
}
