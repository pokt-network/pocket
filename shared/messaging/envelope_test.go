package messaging

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_UnpackMessage_Roundtrip(t *testing.T) {
	someMsg := &DebugMessage{Action: DebugMessageAction_DEBUG_PERSISTENCE_CLEAR_STATE}
	packedMsg, err := PackMessage(someMsg)
	require.NoError(t, err)

	unpackedMsg, err := UnpackMessage[*DebugMessage](packedMsg)
	require.NoError(t, err)

	require.True(t, proto.Equal(someMsg, unpackedMsg))
}
