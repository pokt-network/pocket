package current_height_provider

import "github.com/pokt-network/pocket/shared/modules"

//go:generate mockgen -source=$GOFILE -destination=../../types/mocks/current_height_provider_mock.go -package=mock_types github.com/pokt-network/pocket/p2p/types CurrentHeightProvider

const ModuleName = "current_height_provider"

type CurrentHeightProvider interface {
	modules.Module

	CurrentHeight() uint64
}
