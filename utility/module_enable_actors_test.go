package utility

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/modules"
	mocks "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/stretchr/testify/assert"
)

func TestEnableActorModules(t *testing.T) {
	tests := []struct {
		name                 string
		config               *configs.Config
		expectedError        string
		expectedNames        []string
		expectedActorModules []modules.Module
		expectedLogMessages  []string
	}{
		{
			name: "servicer only",
			config: &configs.Config{
				Servicer: &configs.ServicerConfig{Enabled: true},
			},
			expectedNames: []string{"servicer"},
		},
		{
			name: "fisherman only",
			config: &configs.Config{
				Fisherman: &configs.FishermanConfig{Enabled: true},
			},
			expectedNames: []string{"fisherman"},
		},
		{
			name: "validator only",
			config: &configs.Config{
				Validator: &configs.ValidatorConfig{Enabled: true},
			},
			expectedNames: []string{"validator"},
		},
		{
			name: "portal only",
			config: &configs.Config{
				Portal: &configs.PortalConfig{Enabled: true},
			},
			expectedNames: []string{"portal"},
		},
		{
			name: "validator and servicer",
			config: &configs.Config{
				Validator: &configs.ValidatorConfig{Enabled: true},
				Servicer:  &configs.ServicerConfig{Enabled: true},
			},
			expectedNames: []string{"validator", "servicer"},
		},
		{
			name: "multiple actors not allowed",
			config: &configs.Config{
				Validator: &configs.ValidatorConfig{Enabled: true},
				Fisherman: &configs.FishermanConfig{Enabled: true},
			},
			expectedError: ErrInvalidActorsEnabled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunMgr := mocks.NewMockRuntimeMgr(ctrl)
			cfg, err := configs.CreateTempConfig(test.config)

			mockRunMgr.EXPECT().GetConfig().Return(cfg).AnyTimes()

			bus, err := runtime.CreateBus(mockRunMgr)

			// Call enableActorModules with the test config
			m, err := Create(bus)

			// Verify error output
			if test.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}

			um, ok := m.(modules.UtilityModule)
			assert.True(t, ok)

			// Verify actor modules
			for _, expectedName := range test.expectedNames {
				module, err := um.GetBus().GetModulesRegistry().GetModule(expectedName)
				assert.NoError(t, err)
				assert.NotNil(t, module)
			}
			assert.Equal(t, len(test.expectedNames), len(um.GetActorModules()))

		})
	}
}
