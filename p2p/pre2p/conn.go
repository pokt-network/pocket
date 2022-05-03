package pre2p

import (
	"fmt"
	"io/ioutil"
	"net"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	"github.com/pokt-network/pocket/shared/config"
)

const (
	TCPNetworkLayerProtocol = "tcp4"
)

func CreateListener(cfg *config.Pre2PConfig) (typesPre2P.TransportLayerConn, error) {
	switch cfg.ConnectionType {
	case config.TCPConnection:
		return createTCPListener(cfg)
	case config.PipeConnection:
		return createPipeListener(cfg)
	default:
		return nil, fmt.Errorf("unknown connection type: %s", cfg.ConnectionType)
	}
}

func CreateDialer(cfg *config.Pre2PConfig, url string) (typesPre2P.TransportLayerConn, error) {
	switch cfg.ConnectionType {
	case config.TCPConnection:
		return createTCPDialer(cfg, url)
	case config.PipeConnection:
		return createPipeDialer(cfg, url)
	default:
		return nil, fmt.Errorf("unknown connection type: %s", cfg.ConnectionType)
	}
}

var _ typesPre2P.TransportLayerConn = &tcpConn{}

type tcpConn struct {
	address  *net.TCPAddr
	listener *net.TCPListener
}

func createTCPListener(cfg *config.Pre2PConfig) (*tcpConn, error) {
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

func createTCPDialer(cfg *config.Pre2PConfig, url string) (*tcpConn, error) {
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

	data, err := ioutil.ReadAll(conn)
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

var _ typesPre2P.TransportLayerConn = &pipeConn{}

type pipeConn struct {
	readConn  net.Conn
	writeConn net.Conn
	addr      *string
}

func createPipeListener(cfg *config.Pre2PConfig) (typesPre2P.TransportLayerConn, error) {
	r, w := net.Pipe()
	return &pipeConn{r, w, nil}, nil
}

func createPipeDialer(cfg *config.Pre2PConfig, url string) (typesPre2P.TransportLayerConn, error) {
	r, w := net.Pipe()
	return &pipeConn{r, w, &url}, nil
}

func (c *pipeConn) IsListener() bool {
	panic("Not implemented")
}

func (c *pipeConn) Read() ([]byte, error) {
	panic("Not implemented")
}

func (c *pipeConn) Write(data []byte) error {
	_, err := c.writeConn.Write(data)
	return err
}

func (c *pipeConn) Close() error {
	if err := c.writeConn.Close(); err != nil {
		return err
	}
	if err := c.readConn.Close(); err != nil {
		return err
	}
	return nil
}
