package e2e_tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

func TestHotstuff_4Nodes1BlockHappyPath(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	// Test configs
	runtimeMgrs := generateNodeRuntimeMgrs(t, numValidators, clockMock)
	buses := generateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	sharedNetworkChannel := make(modules.EventsChannel, 100)
	pocketNodes := createTestConsensusPocketNodes(t, buses, sharedNetworkChannel)
	err := startAllTestPocketNodes(t, pocketNodes)
	require.NoError(t, err)

	// Wait for nodes to reach height=1 by generating a block
	block := WaitForNextBlock(t, clockMock, sharedNetworkChannel, pocketNodes, 1, 0, 500, true)
	require.Equal(t, uint64(1), block.BlockHeader.Height)

	// Expecting NewRound messages for height=2 to be sent after a block is committed
	_, err = waitForProposalMsgs(t, clockMock, sharedNetworkChannel, pocketNodes, 2, uint8(consensus.NewRound), 0, 0, numValidators*numValidators, 500, true)
	require.NoError(t, err)

	// TODO(#615): Add QC verification here after valid block mocking is implemented with issue #352.
	// Test state synchronisation's get block functionality
	// At this stage, first block is persisted, get block request for block height 1 must return non-nill block
	serverNode := pocketNodes[1]

	// Choose node 2 as the requester node
	requesterNode := pocketNodes[2]
	requesterNodePeerAddress := requesterNode.GetBus().GetConsensusModule().GetNodeAddress()

	// Send get block request to the server node
	stateSyncGetBlockMsg := prepareStateSyncGetBlockMessage(t, requesterNodePeerAddress, 1)
	send(t, serverNode, stateSyncGetBlockMsg)

	// Server node is waiting for the get block response message
	receivedMsg, err := waitForNetworkStateSyncEvents(t, clockMock, sharedNetworkChannel, "error waiting for StateSync.GetBlockRequest message", 1, 500, false, reflect.TypeOf(&typesCons.StateSyncMessage_GetBlockRes{}))
	require.NoError(t, err)

	// Verify that it was a get block request of the right height
	msg, err := codec.GetCodec().FromAny(receivedMsg[0])
	require.NoError(t, err)
	stateSyncGetBlockResMessage, ok := msg.(*typesCons.StateSyncMessage)
	require.True(t, ok)
	getBlockRes := stateSyncGetBlockResMessage.GetGetBlockRes()
	require.NotEmpty(t, getBlockRes)

	// Validate the data in the block received
	require.Equal(t, uint64(1), getBlockRes.Block.GetBlockHeader().Height)
}

// TODO: Implement these tests and use them as a starting point for new ones. Consider using ChatGPT to help you out :)

// TODO(#615): Implement this test
func TestQuorumCertificate_Valid(t *testing.T) {
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_InsufficientSignature(t *testing.T) {
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_SignatureFromInvalidValidatorSet(t *testing.T) {
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_SignatureFromJailedValidators(t *testing.T) {
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_SignatureFromUnJailedValidators_Valid(t *testing.T) {
	// Unjailed validators should be able to sign a valid QC.
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_SignatureFromValidAndInvalidValidatorSet(t *testing.T) {
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_QuorumCertificateIsModified(t *testing.T) {
	// Leader modifies the QC after sending the proposal, therefore sent QC is invalid.
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_InvalidSignaturesFromValidValidatorSet(t *testing.T) {
	t.Skip()
}

// DISCUSS: This test scenario is currently more exploratory, and it may or may not be relevant.
func TestQuorumCertificate_ResistenceToSignatureMalleability(t *testing.T) {
	t.Skip()
}

func TestHotstuff_4Nodes1Byzantine1Block(t *testing.T) {
	t.Skip()
}

func TestHotstuff_4Nodes2Byzantine1Block(t *testing.T) {
	t.Skip()
}

func TestHotstuff_4Nodes1BlockNetworkPartition(t *testing.T) {
	t.Skip()
}

func TestHotstuff_4Nodes1Block4Rounds(t *testing.T) {
	t.Skip()
}
func TestHotstuff_4Nodes2Blocks(t *testing.T) {
	t.Skip()
}

func TestHotstuff_4Nodes2NewNodes1Block(t *testing.T) {
	t.Skip()
}

func TestHotstuff_4Nodes2DroppedNodes1Block(t *testing.T) {
	t.Skip()
}

func TestHotstuff_4NodesFailOnPrepare(t *testing.T) {
	t.Skip()
}

func TestHotstuff_4NodesFailOnPrecommit(t *testing.T) {
	t.Skip()
}

func TestHotstuff_4NodesFailOnCommit(t *testing.T) {
	t.Skip()
}

func TestHotstuff_4NodesFailOnDecide(t *testing.T) {
	t.Skip()
}

func TestHotstuff_ValidatorWithLockedQC(t *testing.T) {
	t.Skip()
}

func TestHotstuff_ValidatorWithLockedQCMissingNewRoundMsg(t *testing.T) {
	t.Skip()
}
