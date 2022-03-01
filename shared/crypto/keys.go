package crypto

import "encoding/json"

// TODO(discuss): Consider create a type for signature and having constraints for each type as well.

type Address [AddressLen]byte

type PublicKey interface {
	Bytes() []byte
	String() string
	Address() Address
	Equals(other PublicKey) bool
	VerifyBytes(msg []byte, sig []byte) bool // TODO(andrew): consider renaming to Verify
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
}

func (a *Address) UnmarshalJSON(data []byte) error {
	var address string
	err := json.Unmarshal(data, &address)
	if err != nil {
		return err
	}
	copy(a[:], address)
	return nil
}
