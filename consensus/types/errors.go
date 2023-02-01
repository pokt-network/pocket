package types

// TECHDEBT: Avoid having a centralized file for all errors (harder to maintain and identify).

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"

	"github.com/pokt-network/pocket/shared/codec"
	"google.golang.org/protobuf/proto"
)

// Logs and warnings
const (
	// INFO
	DisregardHotstuffMessage = "Discarding hotstuff message"
	NotLockedOnQC            = "node is not locked on any QC"
	ProposalBlockExtends     = "the ProposalQC block is the same as the LockedQC block"

	// WARN
	NilUtilityContextWarning     = "⚠️ [WARN] utilityContext expected to be nil but is not. TODO: Investigate why this is and fix it"
	InvalidPartialSigInQCWarning = "⚠️ [WARN] QC contains an invalid partial signature"

	// DEBUG
	DebugResetToGenesis  = "🧑‍💻 [DEVELOP] Resetting to genesis..."
	DebugTriggerNextView = "🧑‍💻 [DEVELOP] Triggering next view..."
)

var StepToString map[HotstuffStep]string

func init() {
	StepToString = make(map[HotstuffStep]string, len(HotstuffStep_name))
	for i, step := range HotstuffStep_name {
		StepToString[HotstuffStep(i)] = step
	}
}

// TODO(#288): Improve all of this logging:
// 1. Replace `fmt.Sprintf` with `log.Printf` (or similar)
// 2. Add the appropriate log level (warn, debug, etc...) where appropriate
// 3. Remove this file and move log related text into place (easier to maintain, debug, understand, etc.)

func PacemakerInterrupt(reason string, height uint64, step HotstuffStep, round uint64) string {
	return fmt.Sprintf("⏰ Interrupt ⏰ at (height, step, round): (%d, %s, %d)! Reason: %s", height, StepToString[step], round, reason)
}

func PacemakerTimeout(height uint64, step HotstuffStep, round uint64) string {
	return fmt.Sprintf("⌛ Timed out ⌛ at (height, step, round) (%d, %s, %d)!", height, StepToString[step], round)
}

func PacemakerNewHeight(height uint64) string {
	return fmt.Sprintf("🏁 Starting 1st round 🏁 for height: %d", height)
}

func PacemakerCatchup(height1, step1, round1, height2, step2, round2 uint64) string {
	return fmt.Sprintf("🏃 Pacemaker catching 🏃 up (height, step, round) FROM (%d, %d, %d) TO (%d, %d, %d)", height1, step1, round1, height2, step2, round2)
}

func OptimisticVoteCountWaiting(step HotstuffStep, status string) string {
	return fmt.Sprintf("⏳ Waiting ⏳for more %s messages; %s", StepToString[step], status)
}

func OptimisticVoteCountPassed(height uint64, step HotstuffStep, round uint64) string {
	return fmt.Sprintf("📬 Received enough 📬 votes at (height, step, round) (%d, %s, %d)", height, StepToString[step], round)
}

func CommittingBlock(height uint64, numTxs int) string {
	return fmt.Sprintf("🧱 Committing block 🧱 at height %d with %d transactions", height, numTxs)
}

func ElectedNewLeader(address string, nodeId NodeId, height, round uint64) string {
	return fmt.Sprintf("🙇 Elected leader 🙇 for height/round %d/%d: [%d] (%s)", height, round, nodeId, address)
}

func ElectedSelfAsNewLeader(address string, nodeId NodeId, height, round uint64) string {
	return fmt.Sprintf("👑 I am the leader 👑 for height/round %d/%d: [%d] (%s)", height, round, nodeId, address)
}

func SendingMessage(msg *HotstuffMessage, nodeId NodeId) string {
	return fmt.Sprintf("✉️ Sending message ✉️ to %d at (height, step, round) (%d, %d, %d)", nodeId, msg.Height, msg.Step, msg.Round)
}

func SendingStateSyncMessage(nodeId string, height uint64) string {
	return fmt.Sprintf("🔄 Sending State sync message ✉️ to node  %s at height: (%d)  🔄", nodeId, height)
}

func BroadcastingMessage(msg *HotstuffMessage) string {
	return fmt.Sprintf("📣 Broadcasting message 📣 (height, step, round): (%d, %d, %d)", msg.GetHeight(), msg.GetStep(), msg.GetRound())
}

func RestartTimer() string {
	return fmt.Sprintln("Restarting timer")
}

func WarnInvalidPartialSigInQC(address string, nodeId NodeId) string {
	return fmt.Sprintf("%s: from %s (%d)", InvalidPartialSigInQCWarning, address, nodeId)
}

func WarnMissingPartialSig(msg *HotstuffMessage) string {
	return fmt.Sprintf("⚠️ [WARN] No partial signature found for step %s which should not happen...", StepToString[msg.GetStep()])
}

func WarnDiscardHotstuffMessage(_ *HotstuffMessage, reason string) string {
	return fmt.Sprintf("⚠️ [WARN] %s because: %s", DisregardHotstuffMessage, reason)
}

func WarnUnexpectedMessageInPool(_ *HotstuffMessage, height uint64, step HotstuffStep, round uint64) string {
	return fmt.Sprintf("⚠️ [WARN] Message in pool does not match (height, step, round) of QC being generated; %d, %s, %d", height, StepToString[step], round)
}

func WarnIncompletePartialSig(ps *PartialSignature, msg *HotstuffMessage) string {
	return fmt.Sprintf("⚠️ [WARN] Partial signature is incomplete for step %s which should not happen...", StepToString[msg.GetStep()])
}

func DebugTogglePacemakerManualMode(mode string) string {
	return fmt.Sprintf("🔎 [DEBUG] Toggling pacemaker manual mode to %s", mode)
}

func DebugNodeState(state ConsensusNodeState) string {
	return fmt.Sprintf("🔎 [DEBUG] Node %d is at (Height, Step, Round): (%d, %d, %d)", state.NodeId, state.Height, state.Step, state.Round)
}

// TODO(olshansky): Add source and destination NodeId of message here
func DebugReceivedHandlingHotstuffMessage(msg *HotstuffMessage) string {
	return fmt.Sprintf("🔎 [DEBUG] Received hotstuff msg at (Height, Step, Round): (%d, %d, %d)", msg.Height, msg.GetStep(), msg.Round)
}

// TODO(olshansky): Add source and destination NodeId of message here
func DebugHandlingHotstuffMessage(msg *HotstuffMessage) string {
	return fmt.Sprintf("🔎 [DEBUG] Handling hotstuff msg at (Height, Step, Round): (%d, %d, %d)", msg.Height, msg.GetStep(), msg.Round)
}

// Errors
const (
	nilBLockError                               = "block is nil"
	blockExistsError                            = "block exists but should be nil"
	nilBLockProposalError                       = "block should never be nil when creating a proposal message"
	nilBLockVoteError                           = "block should never be nil when creating a vote message for a proposal"
	proposalNotValidInPrepareError              = "proposal is not valid in the PREPARE step"
	nilQCError                                  = "QC being validated is nil"
	nilQCProposalError                          = "QC should never be nil when creating a proposal message"
	nilBlockInQCError                           = "QC must contain a non nil block"
	nilThresholdSigInQCError                    = "QC must contains a non nil threshold signature"
	notEnoughSignaturesError                    = "did not receive enough partial signature"
	nodeIsLockedOnPastHeightQCError             = "node is locked on a QC from a past height"
	nodeIsLockedOnPastRoundQCError              = "node is locked on a QC from a past round"
	unhandledProposalCaseError                  = "unhandled proposal validation check"
	unnecessaryPartialSigForNewRoundError       = "newRound messages do not need a partial signature"
	unnecessaryPartialSigForLeaderProposalError = "leader proposals do not need a partial signature"
	nilPartialSigError                          = "partial signature cannot be nil"
	nilPartialSigOrSourceNotSpecifiedError      = "partial signature is either nil or source is not specified"
	validatorNotFoundInMapError                 = "trying to verify PartialSignature from Validator but it is not in the validator map"
	invalidPartialSignatureError                = "partial signature on message is invalid"
	olderHeightMessageError                     = "hotstuff message is behind the node's height"
	futureHeightMessageError                    = "hotstuff message is ahead the node's height"
	selfProposalError                           = "hotstuff message is a self proposal"
	olderStepRoundError                         = "hotstuff message is of the right height but from the past"
	unexpectedPacemakerCaseError                = "an unexpected pacemaker case occurred"
	invalidAppHashError                         = "apphash being applied does not equal that from utility"
	byzantineOptimisticThresholdError           = "byzantine optimistic threshold not met"
	consensusMempoolFullError                   = "mempool is full"
	applyBlockError                             = "could not apply block"
	prepareBlockError                           = "could not prepare block"
	commitBlockError                            = "could not commit block"
	replicaPrepareBlockError                    = "node should not call `prepareBlock` if it is not a leader"
	blockSizeTooLargeError                      = "block size is too large"
	sendMessageError                            = "error sending message"
	broadcastMessageError                       = "error broadcasting message"
	createConsensusMessageError                 = "error creating consensus message"
	createStateSyncMessageError                 = "error creating state sync message"
	anteValidationError                         = "discarding hotstuff message because ante validation failed"
	nilLeaderIdError                            = "attempting to send a message to leader when LeaderId is nil"
	newPersistenceReadContextError              = "error creating new persistence read context"
	persistenceGetAllValidatorsError            = "error getting all validators from persistence"
)

var (
	ErrNilBlock                               = errors.New(nilBLockError)
	ErrBlockExists                            = errors.New(blockExistsError)
	ErrNilBlockProposal                       = errors.New(nilBLockProposalError)
	ErrNilBlockVote                           = errors.New(nilBLockVoteError)
	ErrProposalNotValidInPrepare              = errors.New(proposalNotValidInPrepareError)
	ErrNilQC                                  = errors.New(nilQCError)
	ErrNilQCProposal                          = errors.New(nilQCProposalError)
	ErrNilBlockInQC                           = errors.New(nilBlockInQCError)
	ErrNilThresholdSigInQC                    = errors.New(nilThresholdSigInQCError)
	ErrNotEnoughSignatures                    = errors.New(notEnoughSignaturesError)
	ErrNodeLockedPastHeight                   = errors.New(nodeIsLockedOnPastHeightQCError)
	ErrNodeLockedPastRound                    = errors.New(nodeIsLockedOnPastRoundQCError)
	ErrUnhandledProposalCase                  = errors.New(unhandledProposalCaseError)
	ErrUnnecessaryPartialSigForNewRound       = errors.New(unnecessaryPartialSigForNewRoundError)
	ErrUnnecessaryPartialSigForLeaderProposal = errors.New(unnecessaryPartialSigForLeaderProposalError)
	ErrNilPartialSig                          = errors.New(nilPartialSigError)
	ErrNilPartialSigOrSourceNotSpecified      = errors.New(nilPartialSigOrSourceNotSpecifiedError)
	ErrOlderMessage                           = errors.New(olderHeightMessageError)
	ErrFutureMessage                          = errors.New(futureHeightMessageError)
	ErrSelfProposal                           = errors.New(selfProposalError)
	ErrOlderStepRound                         = errors.New(olderStepRoundError)
	ErrUnexpectedPacemakerCase                = errors.New(unexpectedPacemakerCaseError)
	ErrConsensusMempoolFull                   = errors.New(consensusMempoolFullError)
	ErrApplyBlock                             = errors.New(applyBlockError)
	ErrPrepareBlock                           = errors.New(prepareBlockError)
	ErrCommitBlock                            = errors.New(commitBlockError)
	ErrReplicaPrepareBlock                    = errors.New(replicaPrepareBlockError)
	ErrSendMessage                            = errors.New(sendMessageError)
	ErrBroadcastMessage                       = errors.New(broadcastMessageError)
	ErrCreateConsensusMessage                 = errors.New(createConsensusMessageError)
	ErrCreateStateSyncMessage                 = errors.New(createStateSyncMessageError)
	ErrHotstuffValidation                     = errors.New(anteValidationError)
	ErrNilLeaderId                            = errors.New(nilLeaderIdError)
	ErrNewPersistenceReadContext              = errors.New(newPersistenceReadContextError)
	ErrPersistenceGetAllValidators            = errors.New(persistenceGetAllValidatorsError)
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

func ErrMissingValidator(address string, nodeId NodeId) error {
	return fmt.Errorf("%s: %s (%d)", validatorNotFoundInMapError, address, nodeId)
}

func ErrValidatingPartialSig(senderAddr string, senderNodeId NodeId, msg *HotstuffMessage, pubKey string) error {
	return fmt.Errorf("%s: Sender: %s (%d); Height: %d; Step: %s; Round: %d; SigHash: %s; BlockHash: %s; PubKey: %s",
		invalidPartialSignatureError, senderAddr, senderNodeId, msg.Height, StepToString[msg.GetStep()], msg.Round, string(msg.GetPartialSignature().Signature), protoHash(msg.Block), pubKey)
}

func ErrPacemakerUnexpectedMessageHeight(err error, heightCurrent, heightMessage uint64) error {
	return fmt.Errorf("%s: Current: %d; Message: %d ", err, heightCurrent, heightMessage)
}

func ErrPacemakerUnexpectedMessageStepRound(err error, step HotstuffStep, round uint64, msg *HotstuffMessage) error {
	return fmt.Errorf("%s: Current (step, round): (%s, %d); Message (step, round): (%s, %d)", err, StepToString[step], round, StepToString[msg.GetStep()], msg.Round)
}

func ErrUnknownConsensusMessageType(msg any) error {
	return fmt.Errorf("unknown consensus message type: %v", msg)
}

func ErrUnknownStateSyncMessageType(msg interface{}) error {
	return fmt.Errorf("unknown state sync message type: %v", msg)
}

func ErrCreateProposeMessage(step HotstuffStep) error {
	return fmt.Errorf("could not create a %s Propose message", StepToString[step])
}

func ErrCreateVoteMessage(step HotstuffStep) error {
	return fmt.Errorf("could not create a %s Vote message", StepToString[step])
}

func ErrQCInvalid(step HotstuffStep) error {
	return fmt.Errorf("invalid QC in step %s", StepToString[step])
}

func ErrLeaderElection(msg *HotstuffMessage) error {
	return fmt.Errorf("leader election failed: Validator cannot take part in consensus at height %d round %d", msg.Height, msg.Round)
}

func protoHash(m proto.Message) string {
	b, err := codec.GetCodec().Marshal(m)
	if err != nil {
		log.Fatalf("Could not marshal proto message: %v", err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
