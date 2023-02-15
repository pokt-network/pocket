//go:build !production

package types

//go:generate mockgen -source=$GOFILE -destination=./mocks/libp2p_mock.go

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
)

type Host interface {
	host.Host
}

//type Message interface {
//	// TODO: what?
//}

type Subscription interface {
	Topic() string
	Next(ctx context.Context) (interface{}, error)
}

//type SubOpt interface {
//	// TODO: what?
//}

type Topic interface {
	String() string
	Subscribe(opts ...interface{}) (Subscription, error)
}
