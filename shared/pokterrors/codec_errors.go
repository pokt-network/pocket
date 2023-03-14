package pokterrors

import (
	"errors"
	"fmt"
)

const codecErrorsPrefix = "codec"

const (
	ProtoMarshalError   = "an error occurred marshalling the structure in protobuf"
	ProtoUnmarshalError = "an error occurred unmarshalling the structure in protobuf"
	ProtoNewAnyError    = "an error occurred creating the protobuf any"
	ProtoFromAnyError   = "an error occurred getting the structure from the protobuf any"
)

func NewCodecError(code CodecErrorCode, msg string) Error {
	return &stdErr{
		CodeError: Code(code),
		module:    codecErrorsPrefix,
		error:     errors.New(msg),
	}
}

func CodecErrProtoFromAny(err error) Error {
	return NewCodecError(CodecErrorCode_ProtoFromAnyError, fmt.Sprintf("%s: %s", ProtoFromAnyError, err.Error()))
}

func CodecErrProtoMarshal(err error) Error {
	return NewCodecError(CodecErrorCode_ProtoMarshalError, fmt.Sprintf("%s: %s", ProtoMarshalError, err.Error()))
}

func CodecErrProtoUnmarshal(err error) Error {
	return NewCodecError(CodecErrorCode_ProtoUnmarshalError, fmt.Sprintf("%s: %s", ProtoUnmarshalError, err.Error()))
}

func CodecErrProtoNewAny(err error) Error {
	return NewCodecError(CodecErrorCode_ProtoNewAnyError, fmt.Sprintf("%s: %s", ProtoNewAnyError, err.Error()))
}
