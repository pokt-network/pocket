package debug

import (
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	"github.com/pokt-network/pocket/p2p/transport"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const (
	// ALL_HEIGHTS is a special height that will be used to indicate that the actors are valid for all heights (including future heights)
	ALL_HEIGHTS = -1
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
	var stakedActors []*coreTypes.Actor
	stakedActors, ok := dabp.actorsByHeight[ALL_HEIGHTS]
	if ok {
		return stakedActors
	}
	stakedActors = dabp.actorsByHeight[int64(height)]
	return stakedActors
}

func (dabp *debugAddrBookProvider) GetStakedAddrBookAtHeight(height uint64) (typesP2P.AddrBook, error) {
	stakedActors := dabp.getActorsByHeight(height)
	addrBook, err := addrbook_provider.ActorsToAddrBook(dabp, stakedActors)
	if err != nil {
		return nil, err
	}
	return addrBook, nil
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
