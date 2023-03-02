package transport

import (
	"fmt"
	"io"
	"net"
	"sync"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
	typesConfigs "github.com/pokt-network/pocket/runtime/configs/types"
)

const (
	TCPNetworkLayerProtocol = "tcp4"
)

func CreateListener(cfg *configs.P2PConfig) (typesP2P.Transport, error) {
	switch cfg.ConnectionType {
	case typesConfigs.ConnectionType_EmptyConnection:
		return createEmptyListener(cfg)
	case typesConfigs.ConnectionType_TCPConnection:
		return createTCPListener(cfg)
	default:
		return nil, fmt.Errorf("unsupported connection type for listener: %v", cfg.ConnectionType)
	}
}

func CreateDialer(cfg *configs.P2PConfig, url string) (typesP2P.Transport, error) {
	switch cfg.ConnectionType {
	case typesConfigs.ConnectionType_EmptyConnection:
		return createEmptyDialer(cfg, url)
	case typesConfigs.ConnectionType_TCPConnection:
		return createTCPDialer(cfg, url)
	default:
		return nil, fmt.Errorf("unsupported connection type for dialer: %v", cfg.ConnectionType)
	}
}

var _ typesP2P.Transport = &tcpConn{}

type tcpConn struct {
	address  *net.TCPAddr
	listener *net.TCPListener

	muConn sync.Mutex
	conn   net.Conn
}

func createTCPListener(cfg *configs.P2PConfig) (*tcpConn, error) {
	addr, err := net.ResolveTCPAddr(TCPNetworkLayerProtocol, fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port))
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

func (c *tcpConn) ReadAll() ([]byte, error) {
	if !c.IsListener() {
		return nil, fmt.Errorf("connection is not a listener")
	}

	conn, err := c.listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("error accepting connection: %v", err)
	}
	defer conn.Close()

	c.muConn.Lock()
	c.conn = conn
	c.muConn.Unlock()

	return io.ReadAll(c)
}

// Read implements the respective member in the io.Reader interface.
// TECHDEBT (SOON OBSOLETE): Read in this implementation is not intended to be
// called directly and will return an error if `tcpConn.conn` is `nil`.
func (c *tcpConn) Read(buf []byte) (int, error) {
	c.muConn.Lock()
	defer c.muConn.Unlock()

	if c.conn == nil {
		return 0, fmt.Errorf("no connection accepted on listener")
	}

	numBz, err := c.conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			return numBz, io.EOF
		}
		return 0, fmt.Errorf("error reading from conn: %v", err)
	}

	return numBz, nil
}

func (c *tcpConn) Write(data []byte) (int, error) {
	if c.IsListener() {
		return 0, fmt.Errorf("connection is a listener")
	}

	client, err := net.DialTCP(TCPNetworkLayerProtocol, nil, c.address)
	if err != nil {
		return 0, err
	}
	defer client.Close()

	return client.Write(data)
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

func (c *emptyConn) ReadAll() ([]byte, error) {
	return nil, nil
}

func (c *emptyConn) Read(data []byte) (int, error) {
	return 0, nil
}

func (c *emptyConn) Write(data []byte) (int, error) {
	return 0, nil
}

func (c *emptyConn) Close() error {
	return nil
}
