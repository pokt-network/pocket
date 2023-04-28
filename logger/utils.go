package logger

import "github.com/rs/zerolog"

// stringLogArrayMarshaler implements the `zerolog.LogArrayMarshaler` interface
// to marshal an array of strings for use with zerolog.
type StringLogArrayMarshaler struct {
	Strings []string
}

// MarshalZerologArray implements the respective `zerolog.LogArrayMarshaler`
// interface member.
func (marshaler StringLogArrayMarshaler) MarshalZerologArray(arr *zerolog.Array) {
	for _, str := range marshaler.Strings {
		arr.Str(str)
	}
}
