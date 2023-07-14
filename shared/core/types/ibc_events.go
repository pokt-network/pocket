package types

func NewAttribute(key, value []byte) *Attribute {
	return &Attribute{Key: key, Value: value}
}
