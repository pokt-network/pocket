package go_cypherdsl

type Cypherize interface {
	ToCypher() (string, error)
}
