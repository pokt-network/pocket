package messaging

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
)

func Test_UnpackMessage_Roundtrip(t *testing.T) {
	someMsg := &DebugMessage{Action: DebugMessageAction_DEBUG_CLEAR_STATE}
	packedMsg, err := PackMessage(someMsg)
	require.NoError(t, err)

	unpackedMsg, err := UnpackMessage[*DebugMessage](packedMsg)
	require.NoError(t, err)

	if !proto.Equal(someMsg, unpackedMsg) {
		t.Fail()
	}
}
