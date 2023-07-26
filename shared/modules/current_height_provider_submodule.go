package modules

//go:generate mockgen -package=mock_modules -destination=./mocks/current_height_provider_submodule_mock.go github.com/pokt-network/pocket/shared/modules CurrentHeightProvider

const CurrentHeightProviderSubmoduleName = "current_height_provider"

type CurrentHeightProvider interface {
	Submodule

	CurrentHeight() uint64
}

type CurrentHeightProviderOption func(CurrentHeightProvider)
