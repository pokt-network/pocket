package types

type Codec interface {
	Register(interface{}, interface{}, interface{}) (uint16, error)
	Encode(interface{}) ([]byte, error)
	Decode([]byte) (interface{}, error)
}
