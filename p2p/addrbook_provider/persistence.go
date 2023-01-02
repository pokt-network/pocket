package addrbook_provider

import (
	"fmt"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/transport"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ modules.IntegratableModule = &persistenceAddrBookProvider{}
	_ typesP2P.AddrBookProvider  = &persistenceAddrBookProvider{}
)

type persistenceAddrBookProvider struct {
	bus         modules.Bus
	p2pCfg      modules.P2PConfig
	connFactory typesP2P.ConnectionFactory
}

func NewPersistenceAddrBookProvider(bus modules.Bus, p2pCfg modules.P2PConfig, options ...func(*persistenceAddrBookProvider)) *persistenceAddrBookProvider {
	pabp := &persistenceAddrBookProvider{
		bus:         bus,
		p2pCfg:      p2pCfg,
		connFactory: transport.CreateDialer, // default connection factory, overridable with WithConnectionFactory()
	}

	for _, o := range options {
		o(pabp)
	}

	return pabp
}

// WithConnectionFactory allows the user to specify a custom connection factory
func WithConnectionFactory(connFactory typesP2P.ConnectionFactory) func(*persistenceAddrBookProvider) {
	return func(pabp *persistenceAddrBookProvider) {
		pabp.connFactory = connFactory
	}
}

func (pabp *persistenceAddrBookProvider) GetBus() modules.Bus {
	return pabp.bus
}

func (pabp *persistenceAddrBookProvider) SetBus(bus modules.Bus) {
	pabp.bus = bus
}

func (pabp *persistenceAddrBookProvider) GetStakedAddrBookAtHeight(height uint64) (typesP2P.AddrBook, error) {
	persistenceReadContext, err := pabp.GetBus().GetPersistenceModule().NewReadContext(int64(height))
	if err != nil {
		return nil, err
	}
	stakedActors, err := persistenceReadContext.GetAllStakedActors(int64(height))
	if err != nil {
		return nil, err
	}
	// TODO(#203): refactor `ValidatorMap`
	validatorMap := make(modules.ValidatorMap, len(stakedActors))
	for _, v := range stakedActors {
		validatorMap[v.GetAddress()] = v
	}
	addrBook, err := pabp.ActorsToAddrBook(validatorMap)
	if err != nil {
		return nil, err
	}
	return addrBook, nil
}

func (pabp *persistenceAddrBookProvider) ActorsToAddrBook(actors map[string]modules.Actor) (typesP2P.AddrBook, error) {
	book := make(typesP2P.AddrBook, 0)
	for _, v := range actors {
		networkPeer, err := pabp.ActorToNetworkPeer(v)
		if err != nil {
				logger.Global.Warn().Err(err).Msg("error connecting to validator",)
			continue
		}
		book = append(book, networkPeer)
	}
	return book, nil
}

func (pabp *persistenceAddrBookProvider) ActorToNetworkPeer(v modules.Actor) (*typesP2P.NetworkPeer, error) {
	conn, err := pabp.connFactory(pabp.p2pCfg, v.GetGenericParam()) // service url
	if err != nil {
		return nil, fmt.Errorf("error resolving addr: %v", err)
	}

	pubKey, err := cryptoPocket.NewPublicKey(v.GetPublicKey())
	if err != nil {
		return nil, err
	}

	peer := &typesP2P.NetworkPeer{
		Dialer:     conn,
		PublicKey:  pubKey,
		Address:    pubKey.Address(),
		ServiceUrl: v.GetGenericParam(), // service url
	}

	return peer, nil
}
