package types

type Transport interface {
	IsListener() bool
	Read() ([]byte, error)
	Write([]byte) error
	Close() error
}
