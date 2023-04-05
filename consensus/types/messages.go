package types

// TECHDEBT: Avoid having a centralized file for all errors (harder to maintain and identify).

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"google.golang.org/protobuf/proto"
)

// Logs and warnings
const (
	// INFO
	DisregardHotstuffMessage = "Discarding hotstuff message"
	NotLockedOnQC            = "node is not locked on any QC"
	ProposalBlockExtends     = "the ProposalQC block is the same as the LockedQC block"
	DisregardBlock           = "Discarding block"

	// WARN
	NilUtilityUOWWarning = "⚠️ utilityUnitOfWork expected to be nil but is not."

	// DEBUG
	DebugResetToGenesis  = "🧑‍💻 Resetting to genesis..."
	DebugTriggerNextView = "🧑‍💻 Triggering next view..."
)

var StepToString map[HotstuffStep]string

func init() {
	StepToString = make(map[HotstuffStep]string, len(HotstuffStep_name))
	for i, step := range HotstuffStep_name {
		StepToString[HotstuffStep(i)] = step
	}
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
	invalidStateHashError                       = "stateHash being applied does not equal that from utility"
	byzantineOptimisticThresholdError           = "BFT threshold not met"
	consensusMempoolFullError                   = "mempool is full"
	applyBlockError                             = "could not apply block"
	prepareBlockError                           = "could not prepare block"
	commitBlockError                            = "could not commit block"
	replicaPrepareBlockError                    = "replica should not call `prepareBlock`"
	blockSizeTooLargeError                      = "block size is too large"
	sendMessageError                            = "error sending message"
	broadcastMessageError                       = "error broadcasting message"
	createConsensusMessageError                 = "error creating consensus message"
	createStateSyncMessageError                 = "error creating state sync message"
	noQcInReceivedBlockError                    = "block does not contain a quorum certificate"
	blockRetrievalError                         = "couldn't retrieve the block from persistence module"
	anteValidationError                         = "discarding hotstuff message because ante validation failed"
	nilLeaderIdError                            = "attempting to send a message to leader when LeaderId is nil"
	newPersistenceReadContextError              = "error creating new persistence read context"
	persistenceGetAllValidatorsError            = "error getting all validators from persistence"
	stateTransitionEventSendingError            = "error sending state transition message"
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
	ErrNoQcInReceivedBlock                    = errors.New(noQcInReceivedBlockError)
	ErrBlockRetrievalMessage                  = errors.New(blockRetrievalError)
	ErrHotstuffValidation                     = errors.New(anteValidationError)
	ErrNilLeaderId                            = errors.New(nilLeaderIdError)
	ErrNewPersistenceReadContext              = errors.New(newPersistenceReadContextError)
	ErrPersistenceGetAllValidators            = errors.New(persistenceGetAllValidatorsError)
	ErrSendingStateTransition                 = errors.New(stateTransitionEventSendingError)
)

func ErrInvalidBlockSize(blockSize, maxSize uint64) error {
	return fmt.Errorf("%s: %d bytes VS max of %d bytes", blockSizeTooLargeError, blockSize, maxSize)
}

func ErrInvalidStateHash(blockHeaderHash, stateHash string) error {
	return fmt.Errorf("%s: %s != %s", invalidStateHashError, blockHeaderHash, stateHash)
}

func ErrByzantineThresholdCheck(n int, threshold float64) error {
	return fmt.Errorf("%s: (%d > %.2f?)", byzantineOptimisticThresholdError, n, threshold)
}

func ErrSendingStateTransitionEvent(event coreTypes.StateMachineEvent) error {
	return fmt.Errorf("%s for event: %s ", ErrSendingStateTransition, event)
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
		logger.Global.Fatal().Err(err).Msg("Could not marshal proto message")
	}
	return base64.StdEncoding.EncodeToString(b)
}
