package current_height_provider

//go:generate mockgen -source=$GOFILE -destination=../../types/mocks/current_height_provider_mock.go -package=mock_types github.com/pokt-network/pocket/p2p/types CurrentHeightProvider

type CurrentHeightProvider interface {
	CurrentHeight() uint64
}
