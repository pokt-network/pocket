package types

import (
	"errors"
	"fmt"
)

// Errors logged at the top level in the consensus module
const (
	// TODO(olshansk) port over all of the logging info
	DisregardHotstuffMessage = "Discarding hotstuff message because"
	NotLockedOnQC            = "node is not locked on any QC"
	ProposalBlockExtends     = "the ProposalQC block is the same as the LockedQC block"

	SendMessageError            = "error sending message"
	BroadcastMessageError       = "error broadcasting message"
	CreateConsensusMessageError = "error creating consensus message"
)

// Errors propagated throughout the consensus module
const (
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
	UnexpectedPacemakerCaseError                = "an unexpected pacemaker case occurred"
	replicaPrepareBlockError                    = "node should not call `prepareBlock` if it is not a leader"
	leaderApplyBlockError                       = "node should not call `applyBlock` if it is leader"
	blockSizeTooLargeError                      = "block size is too large"
	invalidAppHashError                         = "apphash being applied does not equal that from utility"
	byzantineOptimisticThresholdError           = "byzantine optimistic threshold not met"
	consensusMempoolFullError                   = "mempool is full"
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
	ErrUnexpectedPacemakerCase                = errors.New(UnexpectedPacemakerCaseError)
	ErrReplicaPrepareBlock                    = errors.New(replicaPrepareBlockError)
	ErrLeaderApplyBLock                       = errors.New(leaderApplyBlockError)
	ErrConsensusMempoolFull                   = errors.New(consensusMempoolFullError)
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
