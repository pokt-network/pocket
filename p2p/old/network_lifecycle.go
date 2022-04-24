package p2p

import (
	"reflect"
	"strings"

	"github.com/pokt-network/pocket/p2p/types"
	cfg "github.com/pokt-network/pocket/shared/config"
)

func (m *p2pModule) validateConfiguration(config *cfg.P2PConfig) error {
	requiredCfgEntries := []string{
		"Protocol",
		"Address",
		"ExternalIp",
		"Peers",
		"MaxInbound",
		"MaxOutbound",
		"WireHeaderLength",
		"BufferSize",
		"TimeoutInMs",
	}

	cfg := reflect.ValueOf(config).Elem()

	for _, name := range requiredCfgEntries {
		field := cfg.FieldByName(name)

		if field == (reflect.Value{}) || field.IsZero() {
			return ErrMissingOrEmptyConfigField(name)
		}
	}

	return nil
}

func (m *p2pModule) configure(config *cfg.P2PConfig) error {
	if err := m.validateConfiguration(config); err != nil {
		return err
	}

	m.config = config

	m.protocol = config.Protocol
	m.address = config.ExternalIp // TODO(derrandz): use pcrypto.Address (config.Address)
	m.externaladdr = config.ExternalIp

	m.peerlist = types.NewPeerlist()

	// TODO(derrandz): this is a hack to get going no more no less
	// This hack tries to achieve addressbook injection behavior
	// it basically takes a slice of ips and turns them into peers
	// and generating consecutive uints for them from i to length(peers)
	// while also self-filtering.
	// this is filler code until the expected behavior is well spec'd
	for i, p := range config.Peers {
		peer := types.NewPeer(uint64(i+1), p)

		peerAddrParts := strings.Split(peer.Addr(), ":")
		myAddrParts := strings.Split(m.address, ":")

		peerPort, myPort := peerAddrParts[1], myAddrParts[1]
		if peerPort == myPort {
			m.id = peer.Id()
		}
		m.peerlist.Add(*peer)
	}

	return nil
}

func (m *p2pModule) initCodec() {
	m.c = types.NewProtoMarshaler()
}

func (m *p2pModule) initSocketPools() {
	socketFactory := func() interface{} {
		sck := NewSocket(m.config.BufferSize, m.config.WireHeaderLength, m.config.TimeoutInMs)
		return interface{}(sck) // TODO(derrandz): remember to change this if you end up using unsafe_ptr instead of interface{}
	}

	m.inbound = types.NewRegistry(m.config.MaxInbound, socketFactory)
	m.outbound = types.NewRegistry(m.config.MaxOutbound, socketFactory)
}

func (m *p2pModule) initialize(config *cfg.P2PConfig) error {
	if err := m.configure(config); err != nil {
		return err
	}

	m.initCodec()
	m.initSocketPools()

	return nil
}

func (m *p2pModule) isReady() <-chan uint {
	return m.ready
}

func (m *p2pModule) close() {
	m.isListening.Store(false)
	m.listener.Close()
	close(m.quit)
	close(m.errored)
	m.waiters.Wait()
}

func (m *p2pModule) error(err error) {
	defer m.err.Unlock()
	m.err.Lock()

	if m.err.error != nil {
		m.err.error = err
	}

	m.errored <- 1
}

func (m *p2pModule) hasErrored() bool {
	defer m.err.Unlock()
	m.err.Lock()

	return m.err.error != nil
}
