package transport

import (
	"fmt"
	"io"
	"net"

	types "github.com/pokt-network/pocket/runtime/configs/types"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
)

const (
	TCPNetworkLayerProtocol = "tcp4"
)

func CreateListener(cfg *configs.P2PConfig) (typesP2P.Transport, error) {
	switch cfg.ConnectionType {
	case types.ConnectionType_EmptyConnection:
		return createEmptyListener(cfg)
	case types.ConnectionType_TCPConnection:
		return createTCPListener(cfg)
	default:
		return nil, fmt.Errorf("unsupported connection type for listener: %v", cfg.ConnectionType)
	}
}

func CreateDialer(cfg *configs.P2PConfig, url string) (typesP2P.Transport, error) {
	switch cfg.ConnectionType {
	case types.ConnectionType_EmptyConnection:
		return createEmptyDialer(cfg, url)
	case types.ConnectionType_TCPConnection:
		return createTCPDialer(cfg, url)
	default:
		return nil, fmt.Errorf("unsupported connection type for dialer: %v", cfg.ConnectionType)
	}
}

var _ typesP2P.Transport = &tcpConn{}

type tcpConn struct {
	address  *net.TCPAddr
	listener *net.TCPListener
}

func createTCPListener(cfg *configs.P2PConfig) (*tcpConn, error) {
	addr, err := net.ResolveTCPAddr(TCPNetworkLayerProtocol, fmt.Sprintf(":%d", cfg.ConsensusPort))
	if err != nil {
		return nil, err
	}
	l, err := net.ListenTCP(TCPNetworkLayerProtocol, addr)
	if err != nil {
		return nil, err
	}
	return &tcpConn{
		address:  addr,
		listener: l,
	}, nil
}

func createTCPDialer(_ *configs.P2PConfig, url string) (*tcpConn, error) {
	addr, err := net.ResolveTCPAddr(TCPNetworkLayerProtocol, url)
	if err != nil {
		return nil, err
	}
	return &tcpConn{
		address: addr,
	}, nil
}

func (c *tcpConn) IsListener() bool {
	return c.listener != nil
}

func (c *tcpConn) Read() ([]byte, error) {
	if !c.IsListener() {
		return nil, fmt.Errorf("connection is not a listener")
	}
	conn, err := c.listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("error accepting connection: %v", err)
	}
	defer conn.Close()

	data, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("error reading from conn: %v", err)
	}

	return data, nil
}

func (c *tcpConn) Write(data []byte) error {
	if c.IsListener() {
		return fmt.Errorf("connection is a listener")
	}

	client, err := net.DialTCP(TCPNetworkLayerProtocol, nil, c.address)
	if err != nil {
		return err
	}
	defer client.Close()

	if _, err = client.Write(data); err != nil {
		return err
	}

	return nil
}

func (c *tcpConn) Close() error {
	if c.IsListener() {
		return c.listener.Close()
	}
	return nil
}

var _ typesP2P.Transport = &emptyConn{}

type emptyConn struct {
}

func createEmptyListener(_ *configs.P2PConfig) (typesP2P.Transport, error) {
	return &emptyConn{}, nil
}

func createEmptyDialer(_ *configs.P2PConfig, _ string) (typesP2P.Transport, error) {
	return &emptyConn{}, nil
}

func (c *emptyConn) IsListener() bool {
	return false
}

func (c *emptyConn) Read() ([]byte, error) {
	return nil, nil
}

func (c *emptyConn) Write(data []byte) error {
	return nil
}

func (c *emptyConn) Close() error {
	return nil
}
