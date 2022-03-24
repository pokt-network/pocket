package types

import "errors"

const (
	// TODO(olshansk) port over all of the logging info
	DisregardHotstuffMessage = "Discarding hotstuff message because"
	NotLockedOnQC            = "node is not locked on any QC"
	ProposalBlockExtends     = "the ProposalQC block is the same as the LockedQC block"
)

const (
	nilBLockError                               = "block is nil"
	proposalNotValidInPrepareError              = "proposal is not valid in the PREPARE step"
	nilQCError                                  = "QC being validated is nil"
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
)

var (
	ErrNilBlock                               = errors.New(nilBLockError)
	ErrProposalNotValidInPrepare              = errors.New(proposalNotValidInPrepareError)
	ErrNilQC                                  = errors.New(nilQCError)
	ErrNilBlockInQC                           = errors.New(nilBlockInQCError)
	ErrNilThresholdSigInQC                    = errors.New(nilThresholdSigInQCError)
	ErrNotEnoughSignatures                    = errors.New(notEnoughSignaturesError)
	ErrNodeIsLockedOnPastQC                   = errors.New(nodeIsLockedOnPastQCError)
	ErrUnhandledProposalCase                  = errors.New(unhandledProposalCaseError)
	ErrUnnecessaryPartialSigForNewRound       = errors.New(unnecessaryPartialSigForNewRoundError)
	ErrUnnecessaryPartialSigForLeaderProposal = errors.New(unnecessaryPartialSigForLeaderProposalError)
	ErrNilPartialSig                          = errors.New(nilPartialSigError)
	ErrNilPartialSigOrSourceNotSpecified      = errors.New(nilPartialSigOrSourceNotSpecifiedError)
	ErrValidatorNotFoundInMap                 = errors.New(validatorNotFoundInMapError)
	ErrInvalidPartialSignature                = errors.New(invalidPartialSignatureError)
	ErrOlderMessage                           = errors.New(olderHeightMessageError)
	ErrFutureMessage                          = errors.New(futureHeightMessageError)
	ErrSelfProposal                           = errors.New(selfProposalError)
	ErrOlderStepRound                         = errors.New(olderStepRoundError)
	ErrPacemakerCatchup                       = errors.New(pacemakerCatchupError)
	ErrUnexpectedPacemakerCase                = errors.New(UnexpectedPacemakerCaseError)
	ErrReplicaPrepareBlock                    = errors.New(replicaPrepareBlockError)
	ErrLeaderApplyBLock                       = errors.New(leaderApplyBlockError)
	ErrBlockSizeTooLarge                      = errors.New(blockSizeTooLargeError)
	ErrInvalidApphash                         = errors.New(invalidAppHashError)
)
