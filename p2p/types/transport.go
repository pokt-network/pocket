package types

import "github.com/pokt-network/pocket/runtime/configs"

//go:generate mockgen -source=$GOFILE -destination=./mocks/transport_mock.go github.com/pokt-network/pocket/p2p/types Transport

type Transport interface {
	IsListener() bool
	ReadAll() ([]byte, error)
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
}

type ConnectionFactory func(cfg *configs.P2PConfig, url string) (Transport, error)
