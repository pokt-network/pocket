package types

//go:generate mockgen -package mock_types -destination=./mocks/host_mock.go github.com/libp2p/go-libp2p/core/host Host
//go:generate mockgen -package mock_types -destination=./mocks/network_mock.go github.com/libp2p/go-libp2p/core/network Network
//go:generate mockgen -package mock_types -destination=./mocks/conn_mock.go github.com/libp2p/go-libp2p/core/network Conn
//go:generate mockgen -package mock_types -destination=./mocks/stream_mock.go github.com/libp2p/go-libp2p/core/network Stream
