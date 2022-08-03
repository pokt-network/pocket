package types

import (
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

// TODO(Andrew) consolidate errors into one file after recovery

const ( // Explain: using these numbers as it fits nicely with the other error codes in the prototype
	CodeEmptyTransactionError Code = 2

	CodeNewAddressFromBytesError Code = 9

	CodeHexDecodeFromStringError     Code = 11
	CodeInvalidHashLengthError       Code = 12
	CodeEmptyNetworkIDError          Code = 13
	CodeEmptyProposerError           Code = 14
	CodeEmptyTimestampError          Code = 15
	CodeInvalidTransactionCountError Code = 16
	CodeEmptyAccountError            Code = 17
	CodeNilPoolError                 Code = 18
	CodeEmptyNameError               Code = 19
	CodeEmptyAddressError            Code = 20
	CodeInvalidAddressLenError       Code = 21

	CodeInvalidAmountError  Code = 23
	CodeProtoMarshalError   Code = 25
	CodeProtoUnmarshalError Code = 26
	CodeProtoNewAnyError    Code = 27

	CodeDuplicateTransactionError Code = 35

	CodeGetAllValidatorsError Code = 37

	CodeStringToBigIntError Code = 40

	CodeUpdateParamError Code = 88

	CodeInitParamsError         Code = 96
	CodeGetAllFishermenError    Code = 97
	CodeGetAllServiceNodesError Code = 98
	CodeGetAllAppsError         Code = 99

	CodeGetAllPoolsError          Code = 107
	CodeGetAllAccountsError       Code = 108
	CodeGetAllParamsError         Code = 109
	CodeInsufficientAmountError   Code = 41
	CodeNegativeAmountError       Code = 118
	CodeNilQuorumCertificateError Code = 119

	EmptyTransactionError = "the transaction is empty"

	StringToBigIntError = "an error occurred converting the string primitive to big.Int, the conversion was unsuccessful with base 10"

	GetAllValidatorsError = "an error occurred getting all validators from the state"

	InvalidAmountError = "the amount field is invalid; cannot be converted to big.Int"

	InvalidAddressLenError = "the length of the address is not valid"
	EmptyAddressError      = "the address field is empty"
	EmptyNameError         = "the name field is empty"
	NilPoolError           = "the pool is nil"
	EmptyAccountError      = "the account is nil"

	NewAddressFromBytesError     = "unable to convert the raw bytes to a valid address"
	InvalidTransactionCountError = "the total transactions are less than the block transactions"
	EmptyTimestampError          = "the timestamp field is empty"
	EmptyProposerError           = "the proposer field is empty"
	EmptyNetworkIDError          = "the network id field is empty"
	InvalidHashLengthError       = "the length of the hash is not the correct size"
	NilQuorumCertificateError    = "the quorum certificate is nil"

	HexDecodeFromStringError = "an error occurred decoding the string into hex bytes"
	ProtoMarshalError        = "an error occurred marshalling the structure in protobuf"
	ProtoUnmarshalError      = "an error occurred unmarshalling the structure in protobuf"
	ProtoNewAnyError         = "an error occurred creating the protobuf any"

	UpdateParamError = "an error occurred updating the parameter"

	InitParamsError           = "an error occurred initializing the params in genesis"
	GetAllFishermenError      = "an error occurred getting all of the fishermen¬"
	GetAllAppsError           = "an error occurred getting all of the apps"
	GetAllServiceNodesError   = "an error occurred getting all of the service nodes"
	GetAllPoolsError          = "an error occurred getting all of the pools"
	GetAllAccountsError       = "an error occurred getting all of the accounts"
	GetAllParamsError         = "an error occurred getting all of the params"
	DuplicateTransactionError = "the transaction is already found in the mempool"
	InsufficientAmountError   = "the account has insufficient funds to complete the operation"
	NegativeAmountError       = "the amount is negative"
)

func ErrDuplicateTransaction() Error {
	return NewError(CodeDuplicateTransactionError, DuplicateTransactionError)
}

func ErrStringToBigInt() Error {
	return NewError(CodeStringToBigIntError, fmt.Sprintf("%s", StringToBigIntError))
}

// TODO: We should pass the account address here so it is easier to debug the issue
func ErrInsufficientAmount() Error {
	return NewError(CodeInsufficientAmountError, fmt.Sprintf("%s", InsufficientAmountError))
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
