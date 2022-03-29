package types

import (
	"errors"
	"fmt"
)

// Logs and warnings
const (
	DisregardHotstuffMessage = "Discarding hotstuff message because"
	NotLockedOnQC            = "node is not locked on any QC"
	ProposalBlockExtends     = "the ProposalQC block is the same as the LockedQC block"

	NilUtilityContextWarning = "[WARN] Why is the node utility context not nil when preparing a new block? Releasing for now..."

	DebugResetToGenesis  = "[DEBUG] Resetting to genesis..."
	DebugTriggerNextView = "[DEBUG] Triggering next view..."
)

func OptimisticVoteCountWaiting(step string, status string) string {
	return fmt.Sprintf("Still waiting for more %s messages; %s", step, status)
}

func OptimisticVoteCountPassed(step string) string {
	return fmt.Sprintf("received enough %s votes!", step)
}

func CommittingBlock(height uint64, numTxs int) string {
	return fmt.Sprintf("🧱🧱🧱 Committing block at height %d with %d transactions 🧱🧱🧱", height, numTxs)
}

func ElectedNewLeader(address string, nodeId uint64) string {
	return fmt.Sprintf("Elected new leader: %s (%d)", address, nodeId)
}

func SendingMessageForStep(step string, nodeId int) string {
	return fmt.Sprintf("Sending %s message to %d", step, nodeId)
}

func BroadcastingMessageForStep(step string) string {
	return fmt.Sprintf("Sending message for %s step", step)
}

func DebugTogglePacemakerManualMode(mode string) string {
	return fmt.Sprintf("[DEBUG] Toggling pacemaker manual mode to %s", mode)
}

func DebugNodeState(state ConsensusNodeState) string {
	return fmt.Sprintf("\t[DEBUG] NODE STATE: Node %d is at (Height, Step, Round): (%d, %d, %d)\n", state.NodeId, state.Height, state.Step, state.Round)
}

// Errors
const (
	// Messages used to create error objects
	nilBLockError                               = "block is nil"
	nilBLockProposalError                       = "block should never be nil when creating a proposal message"
	nilBLockVoteError                           = "block should never be nil when creating a vote message for a proposal"
	proposalNotValidInPrepareError              = "proposal is not valid in the PREPARE step"
	nilQCError                                  = "QC being validated is nil"
	nilQCProposalError                          = "QC should never be nil when creating a proposal message"
	nilBlockInQCError                           = "QC must contain a non nil block"
	nilThresholdSigInQCError                    = "QC must contains a non nil threshold signature"
	notEnoughSignaturesError                    = "did not receive enough partial signature"
	nodeIsLockedOnPastQCError                   = "node is locked on a QC from the past"
	unhandledProposalCaseError                  = "warning: unhandled proposal validation check"
	unnecessaryPartialSigForNewRoundError       = "newRound messages do not need a partial signature"
	unnecessaryPartialSigForLeaderProposalError = "leader proposals do not need a partial signature"
	nilPartialSigError                          = "partial signature cannot be nil"
	nilPartialSigOrSourceNotSpecifiedError      = "partial signature is either nil or source is not specified"
	validatorNotFoundInMapError                 = "trying to verify PartialSignature from Validator but it is not in the validator map; address"
	invalidPartialSignatureError                = "partial signature on message is invalid"
	olderHeightMessageError                     = "hotstuff message is behind the node's height"
	futureHeightMessageError                    = "hotstuff message is ahead the node's height"
	selfProposalError                           = "hotstuff message is a self proposal"
	olderStepRoundError                         = "hotstuff message is of the right height but from the past"
	pacemakerCatchupError                       = "pacemaker catching up the node's (height, step, round)"
	unexpectedPacemakerCaseError                = "an unexpected pacemaker case occurred"
	replicaPrepareBlockError                    = "node should not call `prepareBlock` if it is not a leader"
	leaderErrApplyBlock                         = "node should not call `applyBlock` if it is leader"
	blockSizeTooLargeError                      = "block size is too large"
	invalidAppHashError                         = "apphash being applied does not equal that from utility"
	byzantineOptimisticThresholdError           = "byzantine optimistic threshold not met"
	consensusMempoolFullError                   = "mempool is full"
	applyBlockError                             = "could not apply block"
	prepareBlockError                           = "could not prepare block"
	commitBlockError                            = "could not commit block"
	sendMessageError                            = "error sending message"
	broadcastMessageError                       = "error broadcasting message"
	createConsensusMessageError                 = "error creating consensus message"
	hotstuffAnteValidationError                 = "discarding hotstuff message because ante validation failed"
	nilLeaderIdError                            = "attempting to send a message to leader when LeaderId is nil"
)

var (
	ErrNilBlock                               = errors.New(nilBLockError)
	ErrNilBlockProposal                       = errors.New(nilBLockProposalError)
	ErrNilBlockVote                           = errors.New(nilBLockVoteError)
	ErrProposalNotValidInPrepare              = errors.New(proposalNotValidInPrepareError)
	ErrNilQC                                  = errors.New(nilQCError)
	ErrNilQCProposal                          = errors.New(nilQCProposalError)
	ErrNilBlockInQC                           = errors.New(nilBlockInQCError)
	ErrNilThresholdSigInQC                    = errors.New(nilThresholdSigInQCError)
	ErrNotEnoughSignatures                    = errors.New(notEnoughSignaturesError)
	ErrNodeIsLockedOnPastQC                   = errors.New(nodeIsLockedOnPastQCError)
	ErrUnhandledProposalCase                  = errors.New(unhandledProposalCaseError)
	ErrUnnecessaryPartialSigForNewRound       = errors.New(unnecessaryPartialSigForNewRoundError)
	ErrUnnecessaryPartialSigForLeaderProposal = errors.New(unnecessaryPartialSigForLeaderProposalError)
	ErrNilPartialSig                          = errors.New(nilPartialSigError)
	ErrNilPartialSigOrSourceNotSpecified      = errors.New(nilPartialSigOrSourceNotSpecifiedError)
	ErrOlderMessage                           = errors.New(olderHeightMessageError)
	ErrFutureMessage                          = errors.New(futureHeightMessageError)
	ErrSelfProposal                           = errors.New(selfProposalError)
	ErrOlderStepRound                         = errors.New(olderStepRoundError)
	ErrPacemakerCatchup                       = errors.New(pacemakerCatchupError)
	ErrUnexpectedPacemakerCase                = errors.New(unexpectedPacemakerCaseError)
	ErrReplicaPrepareBlock                    = errors.New(replicaPrepareBlockError)
	ErrLeaderApplyBLock                       = errors.New(leaderErrApplyBlock)
	ErrConsensusMempoolFull                   = errors.New(consensusMempoolFullError)
	ErrApplyBlock                             = errors.New(applyBlockError)
	ErrPrepareBlock                           = errors.New(prepareBlockError)
	ErrCommitBlock                            = errors.New(commitBlockError)
	ErrSendMessage                            = errors.New(sendMessageError)
	ErrBroadcastMessage                       = errors.New(broadcastMessageError)
	ErrCreateConsensusMessage                 = errors.New(createConsensusMessageError)
	ErrHotstuffAnteValidation                 = errors.New(hotstuffAnteValidationError)
	ErrNilLeaderId                            = errors.New(nilLeaderIdError)
)

func ErrInvalidBlockSize(blockSize, maxSize uint64) error {
	return fmt.Errorf("%s: %d bytes VS max of %d bytes", blockSizeTooLargeError, blockSize, maxSize)
}

func ErrInvalidAppHash(blockHeaderHash, appHash string) error {
	return fmt.Errorf("%s: %s != %s", invalidAppHashError, blockHeaderHash, appHash)
}

func ErrByzantineThresholdCheck(n int, threshold float64) error {
	return fmt.Errorf("%s: (%d > %.2f?)", byzantineOptimisticThresholdError, n, threshold)
}

func ErrMissingValidator(address string, nodeId uint64) error {
	return fmt.Errorf("%s: %s (%d)", validatorNotFoundInMapError, address, nodeId)
}

func ErrValidatingPartialSig(senderAddr string, senderNodeId, height, round uint64, step, signature, blockHash, pubKey string) error {
	return fmt.Errorf("%s: Sender: %s (%d); Height: %d; Step: %s; Round: %d; SigHash: %s; BlockHash: %s; PubKey: %s",
		invalidPartialSignatureError, senderAddr, senderNodeId, height, step, round, signature, blockHash, pubKey)
}

func ErrPacemakerUnexpectedMessageHeight(err error, heightCurrent, heightMessage uint64) error {
	return fmt.Errorf("%s: Current: %d; Message: %d ", err, heightCurrent, heightMessage)
}

func ErrPacemakerUnexpectedMessageStepRound(err error, stepCurrent string, roundCurrent uint64, stepMessage string, roundMessage uint64) error {
	return fmt.Errorf("%s: Current (step, round): (%s, %d); Message (step, round): (%s, %d)", err, stepCurrent, roundCurrent, stepMessage, roundMessage)
}

func ErrUnknownConsensusMessageType(var1 interface{}) error {
	return fmt.Errorf("unknown consensus message type: %v", var1)
}

func CreateProposeMessageError(step string) string {
	return fmt.Sprintf("Could not create a %s Propose message", step)
}

func CreateVoteMessageError(step string) string {
	return fmt.Sprintf("Could not create a %s Vote message", step)
}

func QCInvalidError(step string) string {
	return fmt.Sprintf("QC is invalid in the %s step", step)
}
