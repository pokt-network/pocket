package unit_of_work

import (
	"encoding/hex"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func DefaultTestingParams(_ *testing.T) *genesis.Params {
	return test_artifacts.DefaultParams()
}

func TestUtilityUnitOfWork_GetGovParams(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	for _, paramName := range utils.GovParamMetadataKeys {
		if strings.Contains(paramName, "_owner") {
			continue
		}
		defaultValueParam := reflect.ValueOf(defaultParams).MethodByName("Get" + utils.GovParamMetadataMap[paramName].PropertyName).Call([]reflect.Value{})[0].Interface()
		require.NotNil(t, defaultValueParam)
		switch id := govParamTypes[paramName]; id {
		case BIGINT:
			gotParam, err := getGovParam[*big.Int](uow, paramName)
			require.NoError(t, err)
			defaultValue, ok := defaultValueParam.(string)
			require.True(t, ok)
			require.Equal(t, defaultValue, gotParam.String())
		case INT:
			gotParam, err := getGovParam[int](uow, paramName)
			require.NoError(t, err)
			defaultValue, ok := defaultValueParam.(int32)
			require.True(t, ok)
			require.Equal(t, int(defaultValue), gotParam)
		case INT64:
			gotParam, err := getGovParam[int64](uow, paramName)
			require.NoError(t, err)
			defaultValue, ok := defaultValueParam.(int32)
			require.True(t, ok)
			require.Equal(t, int64(defaultValue), gotParam)
		case BYTES:
			gotParam, er := getGovParam[[]byte](uow, paramName)
			require.NoError(t, er)
			defaultValue, ok := defaultValueParam.(string)
			require.True(t, ok)
			defaultBz, err := hex.DecodeString(defaultValue)
			require.NoError(t, err)
			require.Equal(t, defaultBz, gotParam)
		case STRING:
			gotParam, err := getGovParam[string](uow, paramName)
			require.NoError(t, err)
			defaultValue, ok := defaultValueParam.(string)
			require.True(t, ok)
			require.Equal(t, defaultValue, gotParam)
		default:
			t.Fatalf("unhandled parameter type identifier: got %d", id)
		}
	}
}

func TestUtilityUnitOfWork_HandleMessageChangeParameter(t *testing.T) {
	cdc := codec.GetCodec()
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetMissedBlocksBurnPercentage())
	gotParam, err := getGovParam[int](uow, typesUtil.MissedBlocksBurnPercentageParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
	newParamValue := int32(2)
	paramOwnerPK := test_artifacts.DefaultParamsOwner
	any, er := cdc.ToAny(&wrapperspb.Int32Value{
		Value: newParamValue,
	})
	require.NoError(t, er)
	msg := &typesUtil.MessageChangeParameter{
		Owner:          paramOwnerPK.Address(),
		ParameterKey:   typesUtil.MissedBlocksBurnPercentageParamName,
		ParameterValue: any,
	}
	require.NoError(t, uow.handleMessageChangeParameter(msg), "handle message change param")
	gotParam, err = getGovParam[int](uow, typesUtil.MissedBlocksBurnPercentageParamName)
	require.NoError(t, err)
	require.Equal(t, int(newParamValue), gotParam)
}

func TestUtilityUnitOfWork_GetParamOwner(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	for _, paramName := range utils.GovParamMetadataKeys {
		paramOwnerName := utils.GovParamMetadataMap[paramName].ParamOwner
		paramOwner, err := getGovParam[string](uow, paramOwnerName)
		require.NoError(t, err)
		gotParam, err := uow.getParamOwner(paramName)
		require.NoError(t, err)
		require.Equal(t, paramOwner, hex.EncodeToString(gotParam))
	}
}
