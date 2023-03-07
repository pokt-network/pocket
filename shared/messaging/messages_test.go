package messaging_test

import (
	"fmt"
	"testing"

	//"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/messaging"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestPocketEnvelope_GetContentType(t *testing.T) {
	tests := []struct {
		msg             proto.Message
		wantContentType string
	}{
		{
			msg:             &messaging.DebugMessage{},
			wantContentType: messaging.DebugMessageEventType,
		},
		{
			msg:             &messaging.NodeStartedEvent{},
			wantContentType: messaging.NodeStartedEventType,
		},
		{
			msg:             &typesCons.HotstuffMessage{},
			wantContentType: messaging.HotstuffMessageContentType,
		},
		{
			msg:             &typesUtil.TxGossipMessage{},
			wantContentType: messaging.TxGossipMessageContentType,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("GetContentType %T", tt.msg), func(t *testing.T) {
			packedMsg, err := messaging.PackMessage(tt.msg)
			require.NoError(t, err)
			if got := packedMsg.GetContentType(); got != tt.wantContentType {
				t.Errorf("packedMsg.GetContentType() = %v, want %v", got, tt.wantContentType)
			}
		})
	}
}
