package pokterrors

import (
	"encoding/hex"
	"fmt"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

const utilityModuleErrorsPrefix = "utility"

const (
	GetStakedAmountsError             = "an error occurred getting the validator's amount staked"
	SetValidatorStakedAmountError     = "an error occurred setting the validator' amount staked"
	EqualVotesError                   = "the votes are identical and not equivocating"
	UnequalRoundsError                = "the round numbers are not equal"
	UnequalVoteTypesError             = "the vote types are not equal"
	UnequalPublicKeysError            = "the two public keys are not equal"
	GetMissedBlocksError              = "an error occurred getting the missed blocks field"
	DecodeMessageError                = "unable to decode the message"
	NotExistsError                    = "the actor does not exist in the state"
	InvalidServiceURLError            = "the service url is not valid"
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
	GetServicersPerSessionAtError     = "an error occurred getting the servicers per session for height"
	GetServicerCountError             = "an error occurred getting the servicer count"
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
	StringToBigIntError               = "error converting string to big int"
	StringToBigFloatError             = "error converting string to big float"
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
	InitGenesisParamError             = "an error occurred initializing the params in genesis"
	GetAllFishermenError              = "an error occurred getting all of the fishermen¬"
	GetAllAppsError                   = "an error occurred getting all of the apps"
	GetAllServicersError              = "an error occurred getting all of the servicers"
	GetAllPoolsError                  = "an error occurred getting all of the pools"
	GetAllAccountsError               = "an error occurred getting all of the accounts"
	GetAllParamsError                 = "an error occurred getting all of the params"
	DuplicateTransactionError         = "the transaction is already found in the mempool"
	InsufficientAmountError           = "the account has insufficient funds to complete the operation"
	NegativeAmountError               = "the amount is negative"
	UnknownActorTypeError             = "the actor type is not recognized"
	UnknownMessageTypeError           = "the message being by the utility message is not recognized"
	ErrBadMessageError                = "unable to decode the transaction message"
	ErrBadSignatureError              = "the signature of the transaction is invalid"
)

func UtilityErrUnknownParam(paramName string) Error {
	return NewUtilityError(UtilityErrorCode_UnknownParamError, fmt.Sprintf("%s: %s", UnknownParamError, paramName))
}

func UtilityErrUnequalPublicKeys() Error {
	return NewUtilityError(UtilityErrorCode_UnequalPublicKeysError, UnequalPublicKeysError)
}

func UtilityErrEqualVotes() Error {
	return NewUtilityError(UtilityErrorCode_EqualVotesError, EqualVotesError)
}

func UtilityErrUnequalVoteTypes() Error {
	return NewUtilityError(UtilityErrorCode_UnequalVoteTypesError, UnequalVoteTypesError)
}

func UtilityErrUnequalHeights() Error {
	return NewUtilityError(UtilityErrorCode_UnequalHeightsError, UnequalHeightsError)
}

func UtilityErrUnequalRounds() Error {
	return NewUtilityError(UtilityErrorCode_UnequalRoundsError, UnequalRoundsError)
}

func UtilityErrInvalidServiceURL(reason string) Error {
	return NewUtilityError(UtilityErrorCode_InvalidServiceURLError, fmt.Sprintf("%s: %s", InvalidServiceURLError, reason))
}

func UtilityErrSetPauseHeight(err error) Error {
	return NewUtilityError(UtilityErrorCode_SetPauseHeightError, fmt.Sprintf("%s: %s", SetPauseHeightError, err.Error()))
}

func UtilityErrGetServicersPerSessionAt(height int64, err error) Error {
	return NewUtilityError(UtilityErrorCode_GetServicersPerSessionAtError, fmt.Sprintf("%s: %d; %s", GetServicersPerSessionAtError, height, err.Error()))
}

func UtilityErrGetServicerCount(chain string, height int64, err error) Error {
	return NewUtilityError(UtilityErrorCode_GetServicerCountError, fmt.Sprintf("%s: %s/%d %s", GetServicerCountError, chain, height, err.Error()))
}

func UtilityErrEmptyParamKey() Error {
	return NewUtilityError(UtilityErrorCode_EmptyParamKeyError, EmptyParamKeyError)
}

func UtilityErrEmptyParamValue() Error {
	return NewUtilityError(UtilityErrorCode_EmptyParamValueError, EmptyParamValueError)
}

func UtilityErrGetOutputAddress(operator []byte, err error) Error {
	return NewUtilityError(UtilityErrorCode_GetOutputAddressError, fmt.Sprintf("%s: %s; %s", GetOutputAddressError, hex.EncodeToString(operator), err.Error()))
}

func UtilityErrGetHeight(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetHeightError, fmt.Sprintf("%s:%s", GetHeightError, err.Error()))
}

func UtilityErrGetMissedBlocks(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetMissedBlocksError, fmt.Sprintf("%s: %s", GetMissedBlocksError, err.Error()))
}

func UtilityErrGetStakedTokens(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetStakedAmountError, GetStakedAmountsError)
}

func UtilityErrSetValidatorStakedAmount(err error) Error {
	return NewUtilityError(UtilityErrorCode_SetValidatorStakedAmountError, SetValidatorStakedAmountError)
}

func UtilityErrGetExists(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetExistsError, fmt.Sprintf("%s: %s", GetExistsError, err.Error()))
}

func UtilityErrGetStakeAmount(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetStakeAmountError, fmt.Sprintf("%s: %s", GetStakeAmountError, err.Error()))
}

func UtilityErrStakeLess() Error {
	return NewUtilityError(UtilityErrorCode_StakeLessError, StakeLessError)
}

func UtilityErrSetMissedBlocks(err error) Error {
	return NewUtilityError(UtilityErrorCode_SetMissedBlocksError, fmt.Sprintf("%s: %s", SetMissedBlocksError, err.Error()))
}

func UtilityErrUnmarshalTransaction(err error) Error {
	return NewUtilityError(UtilityErrorCode_UnmarshalTransaction, fmt.Sprintf("%s: %s", UnmarshalTransactionError, err))
}

func UtilityErrAlreadyExists() Error {
	return NewUtilityError(UtilityErrorCode_AlreadyExistsError, AlreadyExistsError)
}

func UtilityErrNotExists() Error {
	return NewUtilityError(UtilityErrorCode_NotExistsError, NotExistsError)
}

func UtilityErrNilOutputAddress() Error {
	return NewUtilityError(UtilityErrorCode_NilOutputAddress, NilOutputAddressError)
}

func UtilityErrEmptyRelayChains() Error {
	return NewUtilityError(UtilityErrorCode_EmptyRelayChainsError, EmptyRelayChainsError)
}

func UtilityErrInvalidRelayChainLength(got, expected int) Error {
	return NewUtilityError(UtilityErrorCode_InvalidRelayChainLengthError, InvalidRelayChainLengthError)
}

func UtilityErrEmptyRelayChain() Error {
	return NewUtilityError(UtilityErrorCode_EmptyRelayChainError, EmptyRelayChainError)
}

func UtilityErrMinimumStake() Error {
	return NewUtilityError(UtilityErrorCode_MinimumStakeError, MinimumStakeError)
}

func UtilityErrGetParam(paramName string, err error) Error {
	return NewUtilityError(UtilityErrorCode_GetParamError, fmt.Sprintf("%s: %s, %s", GetParamError, paramName, err.Error()))
}

func UtilityErrUnauthorizedParamChange(owner []byte) Error {
	return NewUtilityError(UtilityErrorCode_UnauthorizedParamChangeError, fmt.Sprintf("%s: %s", UnauthorizedParamChangeError, hex.EncodeToString(owner)))
}

func UtilityErrInvalidSigner(address string) Error {
	return NewUtilityError(UtilityErrorCode_InvalidSignerError, fmt.Sprintf("%s: %s", InvalidSignerError, address))
}

func UtilityErrMaxChains(maxChains int) Error {
	return NewUtilityError(UtilityErrorCode_MaxChainsError, fmt.Sprintf("%s: %d", MaxChainsError, maxChains))
}

func UtilityErrAlreadyPaused() Error {
	return NewUtilityError(UtilityErrorCode_AlreadyPausedError, AlreadyPausedError)
}

func UtilityErrNotPaused() Error {
	return NewUtilityError(UtilityErrorCode_NotPausedError, NotPausedError)
}

func UtilityErrNotReadyToUnpause() Error {
	return NewUtilityError(UtilityErrorCode_NotReadyToUnpauseError, NotReadyToUnpauseError)
}

func UtilityErrUnknownStatus(status int32) Error {
	return NewUtilityError(UtilityErrorCode_InvalidStatusError, fmt.Sprintf("%s: unknown status %d", InvalidStatusError, status))
}

func UtilityErrInvalidStatus(got, expected int32) Error {
	return NewUtilityError(UtilityErrorCode_InvalidStatusError, fmt.Sprintf("%s: %d expected %d", InvalidStatusError, got, expected))
}

func UtilityErrInsert(err error) Error {
	return NewUtilityError(UtilityErrorCode_InsertError, fmt.Sprintf("%s: %s", InsertError, err.Error()))
}

func UtilityErrGetReadyToUnstake(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetReadyToUnstakeError, fmt.Sprintf("%s: %s", GetReadyToUnstakeError, err.Error()))
}

func UtilityErrSetStatusPausedBefore(err error, beforeHeight int64) Error {
	return NewUtilityError(UtilityErrorCode_SetStatusPausedBeforeError, fmt.Sprintf("%s: %d %s", SetStatusPausedBeforeError, beforeHeight, err.Error()))
}

func UtilityErrGetStatus(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetStatusError, fmt.Sprintf("%s: %s", GetStatusError, err.Error()))
}

func UtilityErrGetPauseHeight(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetPauseHeightError, fmt.Sprintf("%s: %s", GetPauseHeightError, err.Error()))
}

func UtilityErrSetUnstakingHeightAndStatus(err error) Error {
	return NewUtilityError(UtilityErrorCode_SetUnstakingHeightAndStatusError, fmt.Sprintf("%s: %s", SetUnstakingHeightAndStatus, err.Error()))
}

func UtilityErrGetLatestHeight(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetLatestHeightError, fmt.Sprintf("%s: %s", GetLatestHeightError, err.Error()))
}

func UtilityErrUnknownMessage(msg any) Error {
	return NewUtilityError(UtilityErrorCode_UnknownMessageError, fmt.Sprintf("%s: %T", UnknownMessageError, msg))
}

func UtilityErrGetAccountAmount(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetAccountAmountError, fmt.Sprintf("%s: %s", GetAccountAmountError, err.Error()))
}

func UtilityErrAddAccountAmount(err error) Error {
	return NewUtilityError(UtilityErrorCode_AddAccountAmountError, fmt.Sprintf("%s: %s", AddAccountAmountError, err.Error()))
}

func UtilityErrAddPoolAmount(name string, err error) Error {
	return NewUtilityError(UtilityErrorCode_AddPoolAmountError, fmt.Sprintf("%s: pool: %s, %s", AddPoolAmountError, name, err.Error()))
}

func UtilityErrSubPoolAmount(name string, err error) Error {
	return NewUtilityError(UtilityErrorCode_SubPoolAmountError, fmt.Sprintf("%s: pool: %s, %s", SubPoolAmountError, name, err.Error()))
}

func UtilityErrSetPoolAmount(name string, err error) Error {
	return NewUtilityError(UtilityErrorCode_SetPoolAmountError, fmt.Sprintf("%s: pool: %s, %s", SetPoolAmountError, name, err.Error()))
}

func UtilityErrSetPool(name string, err error) Error {
	return NewUtilityError(UtilityErrorCode_SetPoolError, fmt.Sprintf("%s: pool: %s, %s", SetPoolError, name, err.Error()))
}

func UtilityErrGetPoolAmount(name string, err error) Error {
	return NewUtilityError(UtilityErrorCode_GetPoolAmountError, fmt.Sprintf("%s: pool: %s, %s", GetPoolAmountError, name, err.Error()))
}

func UtilityErrSetAccountAmount(err error) Error {
	return NewUtilityError(UtilityErrorCode_SetAccountError, fmt.Sprintf("%s, %s", SetAccountError, err.Error()))
}

func UtilityErrInterfaceConversion(got, expected any) Error {
	return NewUtilityError(UtilityErrorCode_InterfaceConversionError, fmt.Sprintf("%s: %T, expected %T", InterfaceConversionError, got, expected))
}

func UtilityErrAppHash(err error) Error {
	return NewUtilityError(UtilityErrorCode_AppHashError, fmt.Sprintf("%s: %s", AppHashError, err.Error()))
}

func UtilityErrGetBlockHash(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetBlockHashError, fmt.Sprintf("%s: %s", GetBlockHashError, err.Error()))
}

func UtilityErrInvalidPublicKeyLen(pubKeyLen int) Error {
	return NewUtilityError(UtilityErrorCode_InvalidPublicKeyLenError, fmt.Sprintf("%s: %s", InvalidPublicKeyLenError, cryptoPocket.ErrInvalidPublicKeyLen(pubKeyLen)))
}

func UtilityErrInvalidNonce() Error {
	return NewUtilityError(UtilityErrorCode_InvalidNonceError, InvalidNonceError)
}

func UtilityErrNewPublicKeyFromBytes(err error) Error {
	return NewUtilityError(UtilityErrorCode_NewPublicKeyFromBytesError, fmt.Sprintf("%s: %s", NewPublicKeyFromBytesError, err.Error()))
}

func UtilityErrInvalidProposerCutPercentage() Error {
	return NewUtilityError(UtilityErrorCode_InvalidProposerCutPercentageError, InvalidProposerCutPercentageError)
}

func UtilityErrMaxEvidenceAge() Error {
	return NewUtilityError(UtilityErrorCode_MaxEvidenceAgeError, MaxEvidenceAgeError)
}

func UtilityErrInvalidBlockHeight() Error {
	return NewUtilityError(UtilityErrorCode_InvalidBlockHeightError, InvalidBlockHeightError)
}

func UtilityErrInvalidEvidenceType() Error {
	return NewUtilityError(UtilityErrorCode_InvalidEvidenceTypeError, InvalidEvidenceTypeError)
}

func UtilityErrExportState(err error) Error {
	return NewUtilityError(UtilityErrorCode_ExportStateError, fmt.Sprintf("%s: %s", ExportStateError, err.Error()))
}

func UtilityErrNewFeeFromString(fee string) Error {
	return NewUtilityError(UtilityErrorCode_NewFeeFromStringError, fmt.Sprintf("%s: %s", NewFeeFromStringError, fee))
}

func UtilityErrEmptyNonce() Error {
	return NewUtilityError(UtilityErrorCode_EmptyNonceError, EmptyNonceError)
}

func UtilityErrEmptyPublicKey() Error {
	return NewUtilityError(UtilityErrorCode_EmptyPublicKeyError, EmptyPublicKeyError)
}

func UtilityErrEmptyHash() Error {
	return NewUtilityError(UtilityErrorCode_EmptyHashError, EmptyHashError)
}

func UtilityErrEmptyAmount() Error {
	return NewUtilityError(UtilityErrorCode_EmptyAmountError, EmptyAmountError)
}

func UtilityErrEmptySignature() Error {
	return NewUtilityError(UtilityErrorCode_EmptySignatureError, EmptySignatureError)
}

func UtilityErrSignatureVerificationFailed() Error {
	return NewUtilityError(UtilityErrorCode_SignatureVerificationFailedError, SignatureVerificationFailedError)
}

func UtilityErrDecodeMessage(err error) Error {
	return NewUtilityError(UtilityErrorCode_DecodeMessageError, fmt.Sprintf("%s: %s", DecodeMessageError, err.Error()))
}

func UtilityErrProtoFromAny(err error) Error {
	return NewUtilityError(UtilityErrorCode_ProtoFromAnyError, fmt.Sprintf("%s: %s", ProtoFromAnyError, err.Error()))
}

func UtilityErrTransactionAlreadyCommitted() Error {
	return NewUtilityError(UtilityErrorCode_TransactionAlreadyCommittedError, TransactionAlreadyCommittedError)
}

func UtilityErrTransactionSign(err error) Error {
	return NewUtilityError(UtilityErrorCode_TransactionSignError, fmt.Sprintf("%s: %s", TransactionSignError, err.Error()))
}

func UtilityErrInvalidParamValue(got, expected any) Error {
	return NewUtilityError(UtilityErrorCode_InvalidParamValueError, fmt.Sprintf("%s: got %T expected %T", InvalidParamValueError, got, expected))
}

func UtilityErrNewSavePoint(err error) Error {
	return NewUtilityError(UtilityErrorCode_NewSavePointError, fmt.Sprintf("%s: %s", NewSavePointError, err.Error()))
}

func UtilityErrRollbackSavePoint(err error) Error {
	return NewUtilityError(UtilityErrorCode_RollbackSavePointError, fmt.Sprintf("%s: %s", RollbackSavePointError, err.Error()))
}

func UtilityErrNewPersistenceContext(err error) Error {
	return NewUtilityError(UtilityErrorCode_NewPersistenceContextError, fmt.Sprintf("%s: %s", NewPersistenceContextError, err.Error()))
}

func UtilityErrGetAppHash(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetAppHashError, fmt.Sprintf("%s: %s", GetAppHashError, err.Error()))
}

func UtilityErrResetContext(err error) Error {
	return NewUtilityError(UtilityErrorCode_ResetContextError, fmt.Sprintf("%s: %s", ResetContextError, err.Error()))
}

func UtilityErrDuplicateSavePoint() Error {
	return NewUtilityError(UtilityErrorCode_DuplicateSavePointError, DuplicateSavePointError)
}

func UtilityErrEmptySavePoints() Error {
	return NewUtilityError(UtilityErrorCode_EmptySavePointsError, EmptySavePointsError)
}

func UtilityErrSavePointNotFound() Error {
	return NewUtilityError(UtilityErrorCode_SavePointNotFoundError, SavePointNotFoundError)
}

func UtilityErrCommitContext(err error) Error {
	return NewUtilityError(UtilityErrorCode_CommitContextError, fmt.Sprintf("%s: %s", CommitContextError, err.Error()))
}

func UtilityErrReleaseContext(err error) Error {
	return NewUtilityError(UtilityErrorCode_ReleaseContextError, fmt.Sprintf("%s: %s", ReleaseContextError, err.Error()))
}

func UtilityErrMissingRequiredArg(value string) error {
	return NewUtilityError(UtilityErrorCode_MissingRequiredArgError, fmt.Sprintf("%s: %s", MissingRequiredArgError, value))
}

func UtilityErrSocketRequestTimedOut(addr string, nonce uint32) error {
	return NewUtilityError(UtilityErrorCode_SocketRequestTimedOutError, fmt.Sprintf("%s: %s, %d", SocketRequestTimedOutError, addr, nonce))

}

func UtilityErrUndefinedSocketType(socketType string) error {
	return NewUtilityError(UtilityErrorCode_UndefinedSocketTypeError, fmt.Sprintf("%s: %s", UndefinedSocketTypeError, socketType))
}

func UtilityErrPeerHangUp(err error) error {
	return NewUtilityError(UtilityErrorCode_PeerHangUpError, fmt.Sprintf("%s: %s", PeerHangUpError, err.Error()))
}

func UtilityErrUnexpected(err error) error {
	return NewUtilityError(UtilityErrorCode_UnexpectedSocketError, fmt.Sprintf("%s: %s", UnexpectedSocketError, err.Error()))
}

func UtilityErrPayloadTooBig(bodyLength, acceptedLength uint) error {
	return NewUtilityError(UtilityErrorCode_PayloadTooBigError, fmt.Sprintf("%s: payload length: %d, accepted length: %d", PayloadTooBigError, bodyLength, acceptedLength))
}

func UtilityErrSocketIOStartFailed(socketType string) error {
	return NewUtilityError(UtilityErrorCode_SocketIOStartFailedError, fmt.Sprintf("%s: (%s socket)", SocketIOStartFailedError, socketType))
}

func UtilityErrDuplicateTransaction() Error {
	return NewUtilityError(UtilityErrorCode_DuplicateTransactionError, DuplicateTransactionError)
}

func UtilityErrStringToBigInt(err error) Error {
	return NewUtilityError(UtilityErrorCode_StringToBigIntError, fmt.Sprintf("%s: %s", StringToBigIntError, err.Error()))
}

func UtilityErrStringToBigFloat(err error) Error {
	return NewUtilityError(UtilityErrorCode_StringToBigFloatError, fmt.Sprintf("%s: %s", StringToBigFloatError, err.Error()))
}

func UtilityErrInsufficientAmount(address string) Error {
	return NewUtilityError(UtilityErrorCode_InsufficientAmountError, fmt.Sprintf("%s: with address %s", InsufficientAmountError, address))
}

func UtilityErrNegativeAmountError() Error {
	return NewUtilityError(UtilityErrorCode_NegativeAmountError, NegativeAmountError)
}

func UtilityErrGetAllValidators(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetAllValidatorsError, fmt.Sprintf("%s: %s", GetAllValidatorsError, err.Error()))
}

func UtilityErrGetAllFishermen(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetAllFishermenError, fmt.Sprintf("%s: %s", GetAllFishermenError, err.Error()))
}

func UtilityErrGetAllApps(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetAllAppsError, fmt.Sprintf("%s: %s", GetAllAppsError, err.Error()))
}

func UtilityErrGetAllServicers(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetAllServicersError, fmt.Sprintf("%s: %s", GetAllServicersError, err.Error()))
}

func UtilityErrGetAllPools(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetAllPoolsError, fmt.Sprintf("%s: %s", GetAllPoolsError, err.Error()))
}

func UtilityErrGetAllAccounts(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetAllAccountsError, fmt.Sprintf("%s: %s", GetAllAccountsError, err.Error()))
}

func UtilityErrGetAllParams(err error) Error {
	return NewUtilityError(UtilityErrorCode_GetAllParamsError, fmt.Sprintf("%s: %s", GetAllParamsError, err.Error()))
}

func UtilityErrHexDecodeFromString(err error) Error {
	return NewUtilityError(UtilityErrorCode_HexDecodeFromStringError, fmt.Sprintf("%s: %s", HexDecodeFromStringError, err.Error()))
}

func UtilityErrEmptyAccount() Error {
	return NewUtilityError(UtilityErrorCode_EmptyAccountError, EmptyAccountError)
}

func UtilityErrEmptyAddress() Error {
	return NewUtilityError(UtilityErrorCode_EmptyAddressError, EmptyAddressError)
}

func UtilityErrInvalidAddressLen(err error) Error {
	return NewUtilityError(UtilityErrorCode_InvalidAddressLenError, fmt.Sprintf("%s: %s", InvalidAddressLenError, err.Error()))
}

func UtilityErrInvalidAmount() Error {
	return NewUtilityError(UtilityErrorCode_InvalidAmountError, InvalidAmountError)
}

func UtilityErrEmptyName() Error {
	return NewUtilityError(UtilityErrorCode_EmptyNameError, EmptyNameError)
}

func UtilityErrNilPool() Error {
	return NewUtilityError(UtilityErrorCode_NilPoolError, NilPoolError)
}

func UtilityErrEmptyNetworkID() Error {
	return NewUtilityError(UtilityErrorCode_EmptyNetworkIDError, EmptyNetworkIDError)
}

func UtilityErrEmptyProposer() Error {
	return NewUtilityError(UtilityErrorCode_EmptyProposerError, EmptyProposerError)
}

func UtilityErrEmptyTimestamp() Error {
	return NewUtilityError(UtilityErrorCode_EmptyTimestampError, EmptyTimestampError)
}

func EmptyTransactionErr() Error {
	return NewUtilityError(UtilityErrorCode_EmptyTransactionError, EmptyTransactionError)
}

func UtilityErrInvalidTransactionCount() Error {
	return NewUtilityError(UtilityErrorCode_InvalidTransactionCountError, InvalidTransactionCountError)
}

func UtilityErrInvalidHashLength(hashLen int) Error {
	return NewUtilityError(UtilityErrorCode_InvalidHashLengthError, fmt.Sprintf("%s: %s", InvalidHashLengthError, cryptoPocket.ErrInvalidHashLen(hashLen)))
}

func UtilityErrNilQuorumCertificate() Error {
	return NewUtilityError(UtilityErrorCode_NilQuorumCertificateError, NilQuorumCertificateError)
}

func UtilityErrNewAddressFromBytes(err error) Error {
	return NewUtilityError(UtilityErrorCode_NewAddressFromBytesError, fmt.Sprintf("%s: %s", NewAddressFromBytesError, err.Error()))
}

// CONSIDERATION: Moving this into the `codec` library could reduce some code bloat
func UtilityErrProtoMarshal(err error) Error {
	return NewUtilityError(UtilityErrorCode_ProtoMarshalError, fmt.Sprintf("%s: %s", ProtoMarshalError, err.Error()))
}

func UtilityErrProtoUnmarshal(err error) Error {
	return NewUtilityError(UtilityErrorCode_ProtoUnmarshalError, fmt.Sprintf("%s: %s", ProtoUnmarshalError, err.Error()))
}

func UtilityErrProtoNewAny(err error) Error {
	return NewUtilityError(UtilityErrorCode_ProtoNewAnyError, fmt.Sprintf("%s: %s", ProtoNewAnyError, err.Error()))
}

func UtilityErrUpdateParam(err error) Error {
	return NewUtilityError(UtilityErrorCode_UpdateParamError, fmt.Sprintf("%s: %s", UpdateParamError, err.Error()))
}

func UtilityErrInitGenesisParams(err error) Error {
	return NewUtilityError(UtilityErrorCode_InitGenesisParamsError, fmt.Sprintf("%s: %s", InitGenesisParamError, err.Error()))
}

func UtilityErrUnknownActorType(actorType string) Error {
	return NewUtilityError(UtilityErrorCode_UnknownActorType, fmt.Sprintf("%s: %s", UnknownActorTypeError, actorType))
}

func UtilityErrUnknownMessageType(messageType any) Error {
	return NewUtilityError(UtilityErrorCode_UnknownMessageType, fmt.Sprintf("%s: %v", UnknownMessageTypeError, messageType))
}

func UtilityErrBadMessage(err error) Error {
	return NewUtilityError(UtilityErrorCode_ErrBadMessage, fmt.Sprintf("%s: %s", ErrBadMessageError, err.Error()))
}

func UtilityErrBadSignature(err error) Error {
	return NewUtilityError(UtilityErrorCode_ErrBadSignature, fmt.Sprintf("%s: %s", ErrBadSignatureError, err.Error()))
}
