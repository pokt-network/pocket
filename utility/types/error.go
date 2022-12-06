package types

import (
	"encoding/hex"
	"errors"
	"fmt"
)

type Error interface {
	Code() Code
	error
}

type StdErr struct {
	CodeError Code
	error
}

func (se StdErr) Error() string {
	return fmt.Sprintf("CODE: %v, ERROR: %s", se.Code(), se.error.Error())
}

func (se StdErr) Code() Code {
	return se.CodeError
}

func NewError(code Code, msg string) Error {
	return StdErr{
		CodeError: code,
		error:     errors.New(msg),
	}
}

type Code float64

const (
	CodeOK                               Code = 0
	CodeEmptyTransactionError            Code = 2
	CodeInvalidSignerError               Code = 3
	CodeDecodeMessageError               Code = 4
	CodeUnmarshalTransaction             Code = 5
	CodeUnknownMessageError              Code = 6
	CodeAppHashError                     Code = 7
	CodeNewPublicKeyFromBytesError       Code = 8
	CodeNewAddressFromBytesError         Code = 9
	CodeSignatureVerificationFailedError Code = 10
	CodeHexDecodeFromStringError         Code = 11
	CodeInvalidHashLengthError           Code = 12
	CodeEmptyNetworkIDError              Code = 13
	CodeEmptyProposerError               Code = 14
	CodeEmptyTimestampError              Code = 15
	CodeInvalidTransactionCountError     Code = 16
	CodeEmptyAccountError                Code = 17
	CodeNilPoolError                     Code = 18
	CodeEmptyNameError                   Code = 19
	CodeEmptyAddressError                Code = 20
	CodeInvalidAddressLenError           Code = 21
	CodeInvalidNonceError                Code = 22
	CodeInvalidAmountError               Code = 23
	CodeProtoMarshalError                Code = 25
	CodeProtoUnmarshalError              Code = 26
	CodeProtoNewAnyError                 Code = 27
	CodeProtoFromAnyError                Code = 28
	CodeNewFeeFromStringError            Code = 29
	CodeEmptyNonceError                  Code = 30
	CodeEmptyPublicKeyError              Code = 31
	CodeEmptySignatureError              Code = 32
	CodeDuplicateTransactionError        Code = 35
	CodeTransactionSignError             Code = 36
	CodeGetAllValidatorsError            Code = 37
	CodeInterfaceConversionError         Code = 38
	CodeGetAccountAmountError            Code = 39
	CodeStringToBigIntError              Code = 40
	CodeInsufficientAmountError          Code = 41
	CodeAddAccountAmountError            Code = 42
	CodeSetAccountError                  Code = 43
	CodeGetParamError                    Code = 44
	CodeMinimumStakeError                Code = 45
	CodeEmptyRelayChainError             Code = 46
	CodeEmptyRelayChainsError            Code = 47
	CodeInvalidRelayChainLengthError     Code = 48
	CodeNilOutputAddress                 Code = 49
	CodeInvalidPublicKeyLenError         Code = 50
	CodeEmptyAmountError                 Code = 51
	CodeMaxChainsError                   Code = 52
	CodeInsertError                      Code = 53
	CodeInvalidStatusError               Code = 54
	CodeAddPoolAmountError               Code = 55
	CodeSubPoolAmountError               Code = 56
	CodeGetStatusError                   Code = 57
	CodeSetUnstakingHeightAndStatusError Code = 58
	CodeGetReadyToUnstakeError           Code = 59
	CodeAlreadyExistsError               Code = 60
	CodeGetExistsError                   Code = 61
	CodeGetLatestHeightError             Code = 62

	CodeGetPauseHeightError               Code = 64
	CodeAlreadyPausedError                Code = 65
	CodeSetPauseHeightError               Code = 66
	CodeNotPausedError                    Code = 67
	CodeNotReadyToUnpauseError            Code = 68
	CodeSetStatusPausedBeforeError        Code = 69
	CodeInvalidServiceUrlError            Code = 70
	CodeNotExistsError                    Code = 71
	CodeGetMissedBlocksError              Code = 72
	CodeEmptyHashError                    Code = 73
	CodeInvalidBlockHeightError           Code = 74
	CodeUnequalPublicKeysError            Code = 75
	CodeUnequalVoteTypesError             Code = 76
	CodeEqualVotesError                   Code = 77
	CodeUnequalRoundsError                Code = 78
	CodeMaxEvidenceAgeError               Code = 79
	CodeGetStakedTokensError              Code = 80
	CodeSetValidatorStakedTokensError     Code = 81
	CodeSetPoolAmountError                Code = 82
	CodeGetPoolAmountError                Code = 83
	CodeInvalidProposerCutPercentageError Code = 84
	CodeUnknownParamError                 Code = 85
	CodeUnauthorizedParamChangeError      Code = 86
	CodeInvalidParamValueError            Code = 87
	CodeUpdateParamError                  Code = 88
	CodeGetServiceNodesPerSessionAtError  Code = 89
	CodeGetBlockHashError                 Code = 90
	CodeGetServiceNodeCountError          Code = 91
	CodeEmptyParamKeyError                Code = 92
	CodeEmptyParamValueError              Code = 93
	CodeGetOutputAddressError             Code = 94
	CodeTransactionAlreadyCommittedError  Code = 95
	CodeInitParamsError                   Code = 96
	CodeGetAllFishermenError              Code = 97
	CodeGetAllServiceNodesError           Code = 98
	CodeGetAllAppsError                   Code = 99
	CodeNewPersistenceContextError        Code = 100
	CodeGetAppHashError                   Code = 101
	CodeNewSavePointError                 Code = 102
	CodeRollbackSavePointError            Code = 103
	CodeResetContextError                 Code = 104
	CodeCommitContextError                Code = 105
	CodeReleaseContextError               Code = 106
	CodeGetAllPoolsError                  Code = 107
	CodeGetAllAccountsError               Code = 108
	CodeGetAllParamsError                 Code = 109
	CodeSetPoolError                      Code = 110
	CodeDuplicateSavePointError           Code = 111
	CodeSavePointNotFoundError            Code = 112
	CodeEmptySavePointsError              Code = 113
	CodeInvalidEvidenceTypeError          Code = 114
	CodeExportStateError                  Code = 115
	CodeUnequalHeightsError               Code = 116
	CodeSetMissedBlocksError              Code = 117
	CodeNegativeAmountError               Code = 118
	CodeNilQuorumCertificateError         Code = 119
	CodeMissingRequiredArgError           Code = 120
	CodeSocketRequestTimedOutError        Code = 121
	CodeUndefinedSocketTypeError          Code = 122
	CodePeerHangUpError                   Code = 123
	CodeUnexpectedSocketError             Code = 124
	CodePayloadTooBigError                Code = 125
	CodeSocketIOStartFailedError          Code = 126
	CodeGetStakeAmountError               Code = 127
	CodeStakeLessError                    Code = 128
	CodeGetHeightError                    Code = 129
	CodeUnknownActorType                  Code = 130
	CodeUnknownMessageType                Code = 131

	GetStakedTokensError              = "an error occurred getting the validator staked tokens"
	SetValidatorStakedTokensError     = "an error occurred setting the validator staked tokens"
	EqualVotesError                   = "the votes are identical and not equivocating"
	UnequalRoundsError                = "the round numbers are not equal"
	UnequalVoteTypesError             = "the vote types are not equal"
	UnequalPublicKeysError            = "the two public keys are not equal"
	GetMissedBlocksError              = "an error occurred getting the missed blocks field"
	DecodeMessageError                = "unable to decode the message"
	NotExistsError                    = "the actor does not exist in the state"
	InvalidServiceUrlError            = "the service url is not valid"
	NotReadyToUnpauseError            = "the actor isn't ready to unpause as the minimum number of blocks hasn't passed since pausing"
	NotPausedError                    = "the actor is not paused"
	SetPauseHeightError               = "an error occurred setting the pause height"
	AlreadyPausedError                = "the actor is already paused"
	GetPauseHeightError               = "an error occurred getting the pause height"
	UnmarshalTransactionError         = "an error occurred decoding the transaction"
	AlreadyExistsError                = "the actor already exists in the state"
	GetExistsError                    = "an error occurred when checking if already exists"
	GetStakeAmountError               = "an error occurred getting the stake amount"
	StakeLessError                    = "the stake amount cannot be less than current amount"
	GetReadyToUnstakeError            = "an error occurred getting the 'ready to unstake' group"
	GetLatestHeightError              = "an error occurred getting the latest height"
	SetUnstakingHeightAndStatus       = "an error occurred setting the unstaking height and status"
	GetStatusError                    = "an error occurred getting the staking status"
	InvalidStatusError                = "the staking status is not valid"
	InsertError                       = "an error occurred inserting into persistence"
	MaxChainsError                    = "the amount chains exceeds the maximum value"
	InvalidPublicKeyLenError          = "the public key length is not valid"
	EmptyAmountError                  = "the amount field is empty"
	NilOutputAddressError             = "the output address is nil"
	InvalidRelayChainLengthError      = "the relay chain id length is invalid"
	EmptyRelayChainError              = "the relay chain id is empty"
	EmptyRelayChainsError             = "the relay chains are nil or empty"
	MinimumStakeError                 = "an error occurred because the amount specified is less than the minimum stake"
	GetParamError                     = "an error occurred getting the parameter"
	SetAccountError                   = "an error occurred setting the account"
	AddAccountAmountError             = "an error occurred adding the amount to the account balance"
	AddPoolAmountError                = "an error occurred adding to the pool"
	SubPoolAmountError                = "an error occurred subtracting from the pool"
	SetPoolAmountError                = "an error occurred setting the pool amount"
	GetPoolAmountError                = "an error occurred getting the pool amount"
	InvalidSignerError                = "the signer of the message is not a proper candidate"
	GetAccountAmountError             = "an error occurred getting the account amount"
	UnknownMessageError               = "the message type is unrecognized"
	AppHashError                      = "an error occurred generating the apphash"
	InvalidNonceError                 = "the nonce field is invalid; cannot be converted to big.Int"
	NewPublicKeyFromBytesError        = "unable to convert the raw bytes to a valid public key"
	SignatureVerificationFailedError  = "the public key / signature combination is not valid for the msg"
	ProtoFromAnyError                 = "an error occurred getting the structure from the protobuf any"
	NewFeeFromStringError             = "the fee string is unable to be converted to a valid base 10 number"
	EmptyNonceError                   = "the nonce in the transaction is empty"
	EmptyPublicKeyError               = "the public key field is empty"
	EmptySignatureError               = "the signature field is empty"
	TransactionSignError              = "an error occurred signing the transaction"
	InterfaceConversionError          = "an error occurred converting the interface to an expected type: "
	SetStatusPausedBeforeError        = "an error occurred setting the actor status that were paused before"
	EmptyHashError                    = "the hash is empty"
	InvalidBlockHeightError           = "the block height field is not valid"
	MaxEvidenceAgeError               = "the evidence is too old to be processed"
	InvalidProposerCutPercentageError = "the proposer cut percentage is larger than 100"
	UnknownParamError                 = "the param name is not found in the acl"
	UnauthorizedParamChangeError      = "unauthorized param change, the signer must be address: "
	InvalidParamValueError            = "the param value is not the expected type"
	GetBlockHashError                 = "an error occurred getting the block hash"
	GetServiceNodesPerSessionAtError  = "an error occurred getting the service nodes per session for height"
	GetServiceNodeCountError          = "an error occurred getting the service node count"
	EmptyParamKeyError                = "the parameter key is empty"
	EmptyParamValueError              = "the parameter value is empty"
	GetOutputAddressError             = "an error occurred getting the output address using operator"
	GetHeightError                    = "an error occurred when getting the height from the store"
	TransactionAlreadyCommittedError  = "the transaction is already committed"
	NewSavePointError                 = "an error occurred creating the save point"
	RollbackSavePointError            = "an error occurred rolling back to save point"
	NewPersistenceContextError        = "an error occurred creating the persistence context"
	GetAppHashError                   = "an error occurred getting the appHash"
	ResetContextError                 = "an error occurred resetting the context"
	CommitContextError                = "an error occurred committing the context"
	ReleaseContextError               = "an error occurred releasing the context"
	SetPoolError                      = "an error occurred setting the pool"
	DuplicateSavePointError           = "the save point is duplicated"
	SavePointNotFoundError            = "the save point is not found"
	EmptySavePointsError              = "the save points list in context is empty"
	InvalidEvidenceTypeError          = "the evidence type is not valid"
	ExportStateError                  = "an error occurred exporting the state"
	UnequalHeightsError               = "the heights are not equal"
	SetMissedBlocksError              = "an error occurred setting missed blocks"
	MissingRequiredArgError           = "socket error: missing required argument."
	SocketRequestTimedOutError        = "socket error: request timed out while waiting on ACK."
	UndefinedSocketTypeError          = "socket error: undefined given socket type."
	PeerHangUpError                   = "socket error: Peer hang up."
	UnexpectedSocketError             = "socket error: Unexpected peer error."
	PayloadTooBigError                = "socket error: payload size is too big. "
	SocketIOStartFailedError          = "socket error: failed to start socket reading/writing (io)"
	EmptyTransactionError             = "the transaction is empty"
	StringToBigIntError               = "an error occurred converting the string primitive to big.Int, the conversion was unsuccessful with base 10"
	GetAllValidatorsError             = "an error occurred getting all validators from the state"
	InvalidAmountError                = "the amount field is invalid; cannot be converted to big.Int"
	InvalidAddressLenError            = "the length of the address is not valid"
	EmptyAddressError                 = "the address field is empty"
	EmptyNameError                    = "the name field is empty"
	NilPoolError                      = "the pool is nil"
	EmptyAccountError                 = "the account is nil"
	NewAddressFromBytesError          = "unable to convert the raw bytes to a valid address"
	InvalidTransactionCountError      = "the total transactions are less than the block transactions"
	EmptyTimestampError               = "the timestamp field is empty"
	EmptyProposerError                = "the proposer field is empty"
	EmptyNetworkIDError               = "the network id field is empty"
	InvalidHashLengthError            = "the length of the hash is not the correct size"
	NilQuorumCertificateError         = "the quorum certificate is nil"
	HexDecodeFromStringError          = "an error occurred decoding the string into hex bytes"
	ProtoMarshalError                 = "an error occurred marshalling the structure in protobuf"
	ProtoUnmarshalError               = "an error occurred unmarshalling the structure in protobuf"
	ProtoNewAnyError                  = "an error occurred creating the protobuf any"
	UpdateParamError                  = "an error occurred updating the parameter"
	InitParamsError                   = "an error occurred initializing the params in genesis"
	GetAllFishermenError              = "an error occurred getting all of the fishermen¬"
	GetAllAppsError                   = "an error occurred getting all of the apps"
	GetAllServiceNodesError           = "an error occurred getting all of the service nodes"
	GetAllPoolsError                  = "an error occurred getting all of the pools"
	GetAllAccountsError               = "an error occurred getting all of the accounts"
	GetAllParamsError                 = "an error occurred getting all of the params"
	DuplicateTransactionError         = "the transaction is already found in the mempool"
	InsufficientAmountError           = "the account has insufficient funds to complete the operation"
	NegativeAmountError               = "the amount is negative"
	UnknownActorTypeError             = "the actor type is not recognized"
	UnknownMessageTypeError           = "the message being by the utility message is not recognized"
)

func ErrUnknownParam(paramName string) Error {
	return NewError(CodeUnknownParamError, fmt.Sprintf("%s: %s", UnknownParamError, paramName))
}

func ErrUnequalPublicKeys() Error {
	return NewError(CodeUnequalPublicKeysError, fmt.Sprintf("%s", UnequalPublicKeysError))
}

func ErrEqualVotes() Error {
	return NewError(CodeEqualVotesError, fmt.Sprintf("%s", EqualVotesError))
}

func ErrUnequalVoteTypes() Error {
	return NewError(CodeUnequalVoteTypesError, fmt.Sprintf("%s", UnequalVoteTypesError))
}

func ErrUnequalHeights() Error {
	return NewError(CodeUnequalHeightsError, fmt.Sprintf("%s", UnequalHeightsError))
}

func ErrUnequalRounds() Error {
	return NewError(CodeUnequalRoundsError, fmt.Sprintf("%s", UnequalRoundsError))
}

func ErrInvalidServiceUrl(reason string) Error {
	return NewError(CodeInvalidServiceUrlError, fmt.Sprintf("%s: %s", InvalidServiceUrlError, reason))
}

func ErrSetPauseHeight(err error) Error {
	return NewError(CodeSetPauseHeightError, fmt.Sprintf("%s: %s", SetPauseHeightError, err.Error()))
}

func ErrGetServiceNodesPerSessionAt(height int64, err error) Error {
	return NewError(CodeGetServiceNodesPerSessionAtError, fmt.Sprintf("%s: %d; %s", GetServiceNodesPerSessionAtError, height, err.Error()))
}

func ErrGetServiceNodeCount(chain string, height int64, err error) Error {
	return NewError(CodeGetServiceNodeCountError, fmt.Sprintf("%s: %s/%d %s", GetServiceNodeCountError, chain, height, err.Error()))
}

func ErrEmptyParamKey() Error {
	return NewError(CodeEmptyParamKeyError, fmt.Sprintf("%s", EmptyParamKeyError))
}

func ErrEmptyParamValue() Error {
	return NewError(CodeEmptyParamValueError, fmt.Sprintf("%s", EmptyParamValueError))
}

func ErrGetOutputAddress(operator []byte, err error) Error {
	return NewError(CodeGetOutputAddressError, fmt.Sprintf("%s: %s; %s", GetOutputAddressError, hex.EncodeToString(operator), err.Error()))
}

func ErrGetHeight(err error) Error {
	return NewError(CodeGetHeightError, fmt.Sprintf("%s:%s", GetHeightError, err.Error()))
}

func ErrGetMissedBlocks(err error) Error {
	return NewError(CodeGetMissedBlocksError, fmt.Sprintf("%s: %s", GetMissedBlocksError, err.Error()))
}

func ErrGetStakedTokens(err error) Error {
	return NewError(CodeGetStakedTokensError, fmt.Sprintf("%s", GetStakedTokensError))
}

func ErrSetValidatorStakedTokens(err error) Error {
	return NewError(CodeSetValidatorStakedTokensError, fmt.Sprintf("%s", SetValidatorStakedTokensError))
}

func ErrGetExists(err error) Error {
	return NewError(CodeGetExistsError, fmt.Sprintf("%s: %s", GetExistsError, err.Error()))
}

func ErrGetStakeAmount(err error) Error {
	return NewError(CodeGetStakeAmountError, fmt.Sprintf("%s: %s", GetStakeAmountError, err.Error()))
}

func ErrStakeLess() Error {
	return NewError(CodeStakeLessError, StakeLessError)
}

func ErrSetMissedBlocks(err error) Error {
	return NewError(CodeSetMissedBlocksError, fmt.Sprintf("%s: %s", SetMissedBlocksError, err.Error()))
}

func ErrUnmarshalTransaction(err error) Error {
	return NewError(CodeUnmarshalTransaction, fmt.Sprintf("%s: %s", UnmarshalTransactionError, err))
}

func ErrAlreadyExists() Error {
	return NewError(CodeAlreadyExistsError, fmt.Sprintf("%s", AlreadyExistsError))
}

func ErrNotExists() Error {
	return NewError(CodeNotExistsError, fmt.Sprintf("%s", NotExistsError))
}

func ErrNilOutputAddress() Error {
	return NewError(CodeNilOutputAddress, fmt.Sprintf("%s", NilOutputAddressError))
}

func ErrEmptyRelayChains() Error {
	return NewError(CodeEmptyRelayChainsError, fmt.Sprintf("%s", EmptyRelayChainsError))
}

func ErrInvalidRelayChainLength(got, expected int) Error {
	return NewError(CodeInvalidRelayChainLengthError, fmt.Sprintf("%s", InvalidRelayChainLengthError))
}

func ErrEmptyRelayChain() Error {
	return NewError(CodeEmptyRelayChainError, fmt.Sprintf("%s", EmptyRelayChainError))
}

func ErrMinimumStake() Error {
	return NewError(CodeMinimumStakeError, fmt.Sprintf("%s", MinimumStakeError))
}

func ErrGetParam(paramName string, err error) Error {
	return NewError(CodeGetParamError, fmt.Sprintf("%s: %s, %s", GetParamError, paramName, err.Error()))
}

func ErrUnauthorizedParamChange(owner []byte) Error {
	return NewError(CodeUnauthorizedParamChangeError, fmt.Sprintf("%s: %s", UnauthorizedParamChangeError, hex.EncodeToString(owner)))
}

func ErrInvalidSigner() Error {
	return NewError(CodeInvalidSignerError, fmt.Sprintf("%s", InvalidSignerError))
}

func ErrMaxChains(maxChains int) Error {
	return NewError(CodeMaxChainsError, fmt.Sprintf("%s: %d", MaxChainsError, maxChains))
}

func ErrAlreadyPaused() Error {
	return NewError(CodeAlreadyPausedError, fmt.Sprintf("%s", AlreadyPausedError))
}

func ErrNotPaused() Error {
	return NewError(CodeNotPausedError, fmt.Sprintf("%s", NotPausedError))
}

func ErrNotReadyToUnpause() Error {
	return NewError(CodeNotReadyToUnpauseError, fmt.Sprintf("%s", NotReadyToUnpauseError))
}

func ErrInvalidStatus(got, expected int32) Error {
	return NewError(CodeInvalidStatusError, fmt.Sprintf("%s: %d expected %d", InvalidStatusError, got, expected))
}

func ErrInsert(err error) Error {
	return NewError(CodeInsertError, fmt.Sprintf("%s: %s", InsertError, err.Error()))
}

func ErrGetReadyToUnstake(err error) Error {
	return NewError(CodeGetReadyToUnstakeError, fmt.Sprintf("%s: %s", GetReadyToUnstakeError, err.Error()))
}

func ErrSetStatusPausedBefore(err error, beforeHeight int64) Error {
	return NewError(CodeSetStatusPausedBeforeError, fmt.Sprintf("%s: %d %s", SetStatusPausedBeforeError, beforeHeight, err.Error()))
}

func ErrGetStatus(err error) Error {
	return NewError(CodeGetStatusError, fmt.Sprintf("%s: %s", GetStatusError, err.Error()))
}

func ErrGetPauseHeight(err error) Error {
	return NewError(CodeGetPauseHeightError, fmt.Sprintf("%s: %s", GetPauseHeightError, err.Error()))
}

func ErrSetUnstakingHeightAndStatus(err error) Error {
	return NewError(CodeSetUnstakingHeightAndStatusError, fmt.Sprintf("%s: %s", SetUnstakingHeightAndStatus, err.Error()))
}

func ErrGetLatestHeight(err error) Error {
	return NewError(CodeGetLatestHeightError, fmt.Sprintf("%s: %s", GetLatestHeightError, err.Error()))
}

func ErrUnknownMessage(msg interface{}) Error {
	return NewError(CodeUnknownMessageError, fmt.Sprintf("%s: %T", UnknownMessageError, msg))
}

func ErrGetAccountAmount(err error) Error {
	return NewError(CodeGetAccountAmountError, fmt.Sprintf("%s: %s", GetAccountAmountError, err.Error()))
}

func ErrAddAccountAmount(err error) Error {
	return NewError(CodeAddAccountAmountError, fmt.Sprintf("%s: %s", AddAccountAmountError, err.Error()))
}

func ErrAddPoolAmount(name string, err error) Error {
	return NewError(CodeAddPoolAmountError, fmt.Sprintf("%s: pool: %s, %s", AddPoolAmountError, name, err.Error()))
}

func ErrSubPoolAmount(name string, err error) Error {
	return NewError(CodeSubPoolAmountError, fmt.Sprintf("%s: pool: %s, %s", SubPoolAmountError, name, err.Error()))
}

func ErrSetPoolAmount(name string, err error) Error {
	return NewError(CodeSetPoolAmountError, fmt.Sprintf("%s: pool: %s, %s", SetPoolAmountError, name, err.Error()))
}

func ErrSetPool(name string, err error) Error {
	return NewError(CodeSetPoolError, fmt.Sprintf("%s: pool: %s, %s", SetPoolError, name, err.Error()))
}

func ErrGetPoolAmount(name string, err error) Error {
	return NewError(CodeGetPoolAmountError, fmt.Sprintf("%s: pool: %s, %s", GetPoolAmountError, name, err.Error()))
}

func ErrSetAccountAmount(err error) Error {
	return NewError(CodeSetAccountError, fmt.Sprintf("%s, %s", SetAccountError, err.Error()))
}

func ErrInterfaceConversion(got interface{}, expected interface{}) Error {
	return NewError(CodeInterfaceConversionError, fmt.Sprintf("%s: %T, expected %T", InterfaceConversionError, got, expected))
}

func ErrAppHash(err error) Error {
	return NewError(CodeAppHashError, fmt.Sprintf("%s: %s", AppHashError, err.Error()))
}

func ErrGetBlockHash(err error) Error {
	return NewError(CodeGetBlockHashError, fmt.Sprintf("%s: %s", GetBlockHashError, err.Error()))
}

func ErrInvalidPublicKeyLen(err error) Error {
	return NewError(CodeInvalidPublicKeyLenError, fmt.Sprintf("%s: %s", InvalidPublicKeyLenError, err.Error()))
}

func ErrInvalidNonce() Error {
	return NewError(CodeInvalidNonceError, InvalidNonceError)
}

func ErrNewPublicKeyFromBytes(err error) Error {
	return NewError(CodeNewPublicKeyFromBytesError, fmt.Sprintf("%s: %s", NewPublicKeyFromBytesError, err.Error()))
}

func ErrInvalidProposerCutPercentage() Error {
	return NewError(CodeInvalidProposerCutPercentageError, fmt.Sprintf("%s", InvalidProposerCutPercentageError))
}

func ErrMaxEvidenceAge() Error {
	return NewError(CodeMaxEvidenceAgeError, fmt.Sprintf("%s", MaxEvidenceAgeError))
}

func ErrInvalidBlockHeight() Error {
	return NewError(CodeInvalidBlockHeightError, InvalidBlockHeightError)
}

func ErrInvalidEvidenceType() Error {
	return NewError(CodeInvalidEvidenceTypeError, InvalidEvidenceTypeError)
}

func ErrExportState(err error) Error {
	return NewError(CodeExportStateError, fmt.Sprintf("%s: %s", ExportStateError, err.Error()))
}

func ErrNewFeeFromString(fee string) Error {
	return NewError(CodeNewFeeFromStringError, fmt.Sprintf("%s: %s", NewFeeFromStringError, fee))
}

func ErrEmptyNonce() Error {
	return NewError(CodeEmptyNonceError, EmptyNonceError)
}

func ErrEmptyPublicKey() Error {
	return NewError(CodeEmptyPublicKeyError, EmptyPublicKeyError)
}

func ErrEmptyHash() Error {
	return NewError(CodeEmptyHashError, EmptyHashError)
}

func ErrEmptyAmount() Error {
	return NewError(CodeEmptyAmountError, EmptyAmountError)
}

func ErrEmptySignature() Error {
	return NewError(CodeEmptySignatureError, EmptySignatureError)
}

func ErrSignatureVerificationFailed() Error {
	return NewError(CodeSignatureVerificationFailedError, SignatureVerificationFailedError)
}

func ErrDecodeMessage() Error {
	return NewError(CodeDecodeMessageError, fmt.Sprintf("%s", DecodeMessageError))
}

func ErrProtoFromAny(err error) Error {
	return NewError(CodeProtoFromAnyError, fmt.Sprintf("%s: %s", ProtoFromAnyError, err.Error()))
}

func ErrTransactionAlreadyCommitted() Error {
	return NewError(CodeTransactionAlreadyCommittedError, TransactionAlreadyCommittedError)
}

func ErrTransactionSign(err error) Error {
	return NewError(CodeTransactionSignError, fmt.Sprintf("%s: %s", TransactionSignError, err.Error()))
}

func ErrInvalidParamValue(got, expected interface{}) Error {
	return NewError(CodeInvalidParamValueError, fmt.Sprintf("%s: got %T expected %T", InvalidParamValueError, got, expected))
}

func ErrNewSavePoint(err error) Error {
	return NewError(CodeNewSavePointError, fmt.Sprintf("%s: %s", NewSavePointError, err.Error()))
}

func ErrRollbackSavePoint(err error) Error {
	return NewError(CodeRollbackSavePointError, fmt.Sprintf("%s: %s", RollbackSavePointError, err.Error()))
}

func ErrNewPersistenceContext(err error) Error {
	return NewError(CodeNewPersistenceContextError, fmt.Sprintf("%s: %s", NewPersistenceContextError, err.Error()))
}

func ErrGetAppHash(err error) Error {
	return NewError(CodeGetAppHashError, fmt.Sprintf("%s: %s", GetAppHashError, err.Error()))
}

func ErrResetContext(err error) Error {
	return NewError(CodeResetContextError, fmt.Sprintf("%s: %s", ResetContextError, err.Error()))
}

func ErrDuplicateSavePoint() Error {
	return NewError(CodeDuplicateSavePointError, fmt.Sprintf("%s", DuplicateSavePointError))
}

func ErrEmptySavePoints() Error {
	return NewError(CodeEmptySavePointsError, fmt.Sprintf("%s", EmptySavePointsError))
}

func ErrSavePointNotFound() Error {
	return NewError(CodeSavePointNotFoundError, fmt.Sprintf("%s", SavePointNotFoundError))
}

func ErrCommitContext(err error) Error {
	return NewError(CodeCommitContextError, fmt.Sprintf("%s: %s", CommitContextError, err.Error()))
}

func ErrReleaseContext(err error) Error {
	return NewError(CodeReleaseContextError, fmt.Sprintf("%s: %s", ReleaseContextError, err.Error()))
}

func ErrMissingRequiredArg(value string) error {
	return NewError(CodeMissingRequiredArgError, fmt.Sprintf("%s: %s", MissingRequiredArgError, value))
}

func ErrSocketRequestTimedOut(addr string, nonce uint32) error {
	return NewError(CodeSocketRequestTimedOutError, fmt.Sprintf("%s: %s, %d", SocketRequestTimedOutError, addr, nonce))

}

func ErrUndefinedSocketType(socketType string) error {
	return NewError(CodeUndefinedSocketTypeError, fmt.Sprintf("%s: %s", UndefinedSocketTypeError, socketType))
}

func ErrPeerHangUp(err error) error {
	return NewError(CodePeerHangUpError, fmt.Sprintf("%s: %s", PeerHangUpError, err.Error()))
}

func ErrUnexpected(err error) error {
	return NewError(CodeUnexpectedSocketError, fmt.Sprintf("%s: %s", UnexpectedSocketError, err.Error()))
}

func ErrPayloadTooBig(bodyLength, acceptedLength uint) error {
	return NewError(CodePayloadTooBigError, fmt.Sprintf("%s: payload length: %d, accepted length: %d", PayloadTooBigError, bodyLength, acceptedLength))
}

func ErrSocketIOStartFailed(socketType string) error {
	return NewError(CodeSocketIOStartFailedError, fmt.Sprintf("%s: (%s socket)", SocketIOStartFailedError, socketType))
}

func ErrDuplicateTransaction() Error {
	return NewError(CodeDuplicateTransactionError, DuplicateTransactionError)
}

func ErrStringToBigInt() Error {
	return NewError(CodeStringToBigIntError, StringToBigIntError)
}

func ErrInsufficientAmount(address string) Error {
	return NewError(CodeInsufficientAmountError, fmt.Sprintf("%s: with address %s", InsufficientAmountError, address))
}

func ErrNegativeAmountError() Error {
	return NewError(CodeNegativeAmountError, fmt.Sprintf("%s", NegativeAmountError))
}

func ErrGetAllValidators(err error) Error {
	return NewError(CodeGetAllValidatorsError, fmt.Sprintf("%s: %s", GetAllValidatorsError, err.Error()))
}

func ErrGetAllFishermen(err error) Error {
	return NewError(CodeGetAllFishermenError, fmt.Sprintf("%s: %s", GetAllFishermenError, err.Error()))
}

func ErrGetAllApps(err error) Error {
	return NewError(CodeGetAllAppsError, fmt.Sprintf("%s: %s", GetAllAppsError, err.Error()))
}

func ErrGetAllServiceNodes(err error) Error {
	return NewError(CodeGetAllServiceNodesError, fmt.Sprintf("%s: %s", GetAllServiceNodesError, err.Error()))
}

func ErrGetAllPools(err error) Error {
	return NewError(CodeGetAllPoolsError, fmt.Sprintf("%s: %s", GetAllPoolsError, err.Error()))
}

func ErrGetAllAccounts(err error) Error {
	return NewError(CodeGetAllAccountsError, fmt.Sprintf("%s: %s", GetAllAccountsError, err.Error()))
}

func ErrGetAllParams(err error) Error {
	return NewError(CodeGetAllParamsError, fmt.Sprintf("%s: %s", GetAllParamsError, err.Error()))
}

func ErrHexDecodeFromString(err error) Error {
	return NewError(CodeHexDecodeFromStringError, fmt.Sprintf("%s: %s", HexDecodeFromStringError, err.Error()))
}

func ErrEmptyAccount() Error {
	return NewError(CodeEmptyAccountError, EmptyAccountError)
}

func ErrEmptyAddress() Error {
	return NewError(CodeEmptyAddressError, EmptyAddressError)
}

func ErrInvalidAddressLen(err error) Error {
	return NewError(CodeInvalidAddressLenError, fmt.Sprintf("%s: %s", InvalidAddressLenError, err.Error()))
}

func ErrInvalidAmount() Error {
	return NewError(CodeInvalidAmountError, InvalidAmountError)
}

func ErrEmptyName() Error {
	return NewError(CodeEmptyNameError, EmptyNameError)
}

func ErrNilPool() Error {
	return NewError(CodeNilPoolError, NilPoolError)
}

func ErrEmptyNetworkID() Error {
	return NewError(CodeEmptyNetworkIDError, EmptyNetworkIDError)
}

func ErrEmptyProposer() Error {
	return NewError(CodeEmptyProposerError, EmptyProposerError)
}

func ErrEmptyTimestamp() Error {
	return NewError(CodeEmptyTimestampError, EmptyTimestampError)
}

func EmptyTransactionErr() Error {
	return NewError(CodeEmptyTransactionError, EmptyTransactionError)
}

func ErrInvalidTransactionCount() Error {
	return NewError(CodeInvalidTransactionCountError, InvalidTransactionCountError)
}

func ErrInvalidHashLength(err error) Error {
	return NewError(CodeInvalidHashLengthError, fmt.Sprintf("%s: %s", InvalidHashLengthError, err.Error()))
}

func ErrNilQuorumCertificate() Error {
	return NewError(CodeNilQuorumCertificateError, NilQuorumCertificateError)
}

func ErrNewAddressFromBytes(err error) Error {
	return NewError(CodeNewAddressFromBytesError, fmt.Sprintf("%s: %s", NewAddressFromBytesError, err.Error()))
}

func ErrProtoMarshal(err error) Error {
	return NewError(CodeProtoMarshalError, fmt.Sprintf("%s: %s", ProtoMarshalError, err.Error()))
}

func ErrProtoUnmarshal(err error) Error {
	return NewError(CodeProtoUnmarshalError, fmt.Sprintf("%s: %s", ProtoUnmarshalError, err.Error()))
}

func ErrProtoNewAny(err error) Error {
	return NewError(CodeProtoNewAnyError, fmt.Sprintf("%s: %s", ProtoNewAnyError, err.Error()))
}

func ErrUpdateParam(err error) Error {
	return NewError(CodeUpdateParamError, fmt.Sprintf("%s: %s", UpdateParamError, err.Error()))
}

func ErrInitParams(err error) Error {
	return NewError(CodeInitParamsError, fmt.Sprintf("%s: %s", InitParamsError, err.Error()))
}

func ErrUnknownActorType(actorType string) Error {
	return NewError(CodeUnknownActorType, fmt.Sprintf("%s: %s", UnknownActorTypeError, actorType))
}

func ErrUnknownMessageType(messageType interface{}) Error {
	return NewError(CodeUnknownMessageType, fmt.Sprintf("%s: %v", UnknownMessageTypeError, messageType))
}
