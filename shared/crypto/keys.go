package crypto

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
)

// TODO(discuss): Consider create a type for signature and having constraints for each type as well.
type Address []byte

type PublicKey interface {
	Bytes() []byte
	String() string
	Address() Address
	Equals(other PublicKey) bool
	Verify(msg []byte, sig []byte) bool
	Size() int
}

type PrivateKey interface {
	Bytes() []byte
	String() string
	Equals(other PrivateKey) bool
	PublicKey() PublicKey
	Address() Address
	Sign(msg []byte) ([]byte, error)
	Size() int
	Seed() []byte
}

func (a *Address) UnmarshalJSON(data []byte) error {
	var address string
	err := json.Unmarshal(data, &address)
	if err != nil {
		return err
	}
	bytes, err := hex.DecodeString(address)
	if err != nil {
		return err
	}
	*a = bytes
	return nil
}

func (a *Address) ToString() string {
	return hex.EncodeToString(*a)
}

func (a Address) Equals(other Address) bool {
	return bytes.Equal(a, other)
}

func AddressFromString(s string) Address {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal("Should never fail on decoding an address from string: ", err)
	}
	return Address(bytes)
}
