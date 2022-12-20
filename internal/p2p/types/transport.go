package types

import "github.com/pokt-network/pocket/internal/shared/modules"

//go:generate mockgen -source=$GOFILE -destination=./mocks/transport_mock.go github.com/pokt-network/pocket/internal/p2p/types Transport

type Transport interface {
	IsListener() bool
	Read() ([]byte, error)
	Write([]byte) error
	Close() error
}

type ConnectionFactory func(cfg modules.P2PConfig, url string) (Transport, error)
