// TODO: This is a PLACEHOLDER implementation of serialization
// using json since its part of the standard lib until we move to protos.

package types

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"log"

	"google.golang.org/protobuf/proto"
)

const Hasher = crypto.SHA256

type Encodable interface {
	ToBytes() []byte
	ToString() string
	Hash() []byte
	HashString() string
}

func GobEncode(m interface{}) ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func GobDecode(data []byte, m interface{}) error {
	var buff = bytes.NewBuffer(data)
	dec := gob.NewDecoder(buff)
	if err := dec.Decode(&m); err != nil {
		return err
	}
	return nil
}

func Hash(b []byte) []byte {
	hasher := Hasher.New()
	hasher.Write(b) //nolint:golint,errcheck
	return hasher.Sum(nil)
}

func HexEncode(b []byte) string {
	return hex.EncodeToString(b)
}

func HashString(b []byte) string {
	return HexEncode(Hash(b))
}

func ProtoHash(m proto.Message) string {
	b, err := proto.Marshal(m)
	if err != nil {
		log.Fatalf("Could not marshal proto message: %v", err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
