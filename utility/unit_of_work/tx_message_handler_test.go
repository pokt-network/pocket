package unit_of_work

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	types "github.com/pokt-network/pocket/shared/core/types"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestHandleMessageUpgrade(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockrwcontext := mockModules.NewMockPersistenceRWContext(ctrl)
	logger := zerolog.New(nil)
	persistenceRWContext := mockrwcontext

	u := &baseUtilityUnitOfWork{
		logger:               &logger,
		persistenceRWContext: persistenceRWContext,
		height:               10,
	}

	// Test: upgrade height must be greater than current height
	msg := &typesUtil.MessageUpgrade{
		Version: "1.0.0",
		Height:  5,
	}
	err := u.handleMessageUpgrade(msg)
	assert.Equal(t, types.ErrSettingUpgrade(errors.New("upgrade height must be greater than current height")), err)

	// Test: invalid protocol version
	msg = &typesUtil.MessageUpgrade{
		Version: "invalid-version",
		Height:  20,
	}
	err = u.handleMessageUpgrade(msg)
	assert.Equal(t, types.ErrInvalidProtocolVersion("invalid-version"), err)

	// Test: error getting current version
	persistenceRWContext.EXPECT().GetVersionAtHeight(gomock.Any()).Return("", errors.New("some error"))
	err = u.handleMessageUpgrade(&typesUtil.MessageUpgrade{Version: "2.0.0", Height: 20})
	assert.Error(t, err)

	// Test: error parsing current version
	persistenceRWContext.EXPECT().GetVersionAtHeight(gomock.Any()).Return("invalid-version", nil)
	err = u.handleMessageUpgrade(&typesUtil.MessageUpgrade{Version: "2.0.0", Height: 20})
	assert.Error(t, err)

	// Test: major version jump too large
	persistenceRWContext.EXPECT().GetVersionAtHeight(gomock.Any()).Return("1.0.0", nil)
	err = u.handleMessageUpgrade(&typesUtil.MessageUpgrade{Version: "3.0.0", Height: 20})
	assert.Error(t, err)

	// Test: version must be greater than current
	persistenceRWContext.EXPECT().GetVersionAtHeight(gomock.Any()).Return("2.0.0", nil)
	err = u.handleMessageUpgrade(&typesUtil.MessageUpgrade{Version: "2.0.0", Height: 20})
	assert.Error(t, err)

	// Test: error setting upgrade
	persistenceRWContext.EXPECT().GetVersionAtHeight(gomock.Any()).Return("1.0.0", nil)
	persistenceRWContext.EXPECT().SetUpgrade(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error"))
	err = u.handleMessageUpgrade(&typesUtil.MessageUpgrade{Version: "2.0.0", Height: 20})
	assert.Error(t, err)

	// Test: successful upgrade
	persistenceRWContext.EXPECT().GetVersionAtHeight(gomock.Any()).Return("1.0.0", nil)
	persistenceRWContext.EXPECT().SetUpgrade(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	err = u.handleMessageUpgrade(&typesUtil.MessageUpgrade{Version: "2.0.0", Height: 20})
	assert.NoError(t, err)
}
