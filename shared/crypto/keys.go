package crypto

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
}
