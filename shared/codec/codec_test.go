package codec

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSharedCodec(t *testing.T) {
	expectedField1One := int32(1)
	expectedField2One := "1"
	expectedField3One := false
	tpsOne := &TestProtoStructure{
		Field1: expectedField1One,
		Field2: expectedField2One,
		Field3: expectedField3One,
	}

	expectedField1Two := int32(2)
	expectedField2Two := "2"
	expectedField3Two := true
	tpsTwo := &TestProtoStructure{
		Field1: expectedField1Two,
		Field2: expectedField2Two,
		Field3: expectedField3Two,
	}

	codec := GetCodec()

	// ensure different test structures
	requireTestProtoStructureNotEqual(t, tpsOne, tpsTwo)

	// test marshalling
	tpsOneProtoBytes, err := codec.Marshal(tpsOne)
	require.NoError(t, err)

	tpsTwoProtoBytes, err := codec.Marshal(tpsTwo)
	require.NoError(t, err)
	require.NotEqual(t, tpsOneProtoBytes, tpsTwoProtoBytes)

	// test unmarshalling
	tpsOneUnmarshalled := &TestProtoStructure{}
	require.NoError(t, err)
	require.NoError(t, codec.Unmarshal(tpsOneProtoBytes, tpsOneUnmarshalled))

	tpsTwoUnmarshalled := &TestProtoStructure{}
	require.NoError(t, err)
	require.NoError(t, codec.Unmarshal(tpsTwoProtoBytes, tpsTwoUnmarshalled))

	requireTestProtoStructureEqual(t, tpsOne, tpsOneUnmarshalled)
	requireTestProtoStructureEqual(t, tpsTwo, tpsTwoUnmarshalled)
	requireTestProtoStructureNotEqual(t, tpsOneUnmarshalled, tpsTwoUnmarshalled)
}

func requireTestProtoStructureEqual(t *testing.T, tpsOne, tpsTwo *TestProtoStructure) {
	require.Equal(t, tpsOne.Field1, tpsTwo.Field1)
	require.Equal(t, tpsOne.Field2, tpsTwo.Field2)
	require.Equal(t, tpsOne.Field3, tpsTwo.Field3)
}

func requireTestProtoStructureNotEqual(t *testing.T, tpsOne, tpsTwo *TestProtoStructure) {
	require.NotEqual(t, tpsOne.Field1, tpsTwo.Field1)
	require.NotEqual(t, tpsOne.Field2, tpsTwo.Field2)
	require.NotEqual(t, tpsOne.Field3, tpsTwo.Field3)
}
