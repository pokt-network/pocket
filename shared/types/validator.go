package types

import pcrypto "github.com/pokt-network/pocket/shared/crypto"

// The key is a hex encoded representation of the validator byte address.
// TODO(discuss): Should there be a type for a stringified version of `Address`?
type ValMap map[string]*Validator

type Validator struct {
	Address   pcrypto.Address          `json:"address"`
	PublicKey pcrypto.Ed25519PublicKey `json:"public_key"`
	Jailed    bool                     `json:"jailed"` // TODO(olshansky): Integrate with utility to update this.
	UPokt     uint64                   `json:"upokt"`  // TODO(olshansky): Integrate with utility to update this.
	Host      string                   `json:"host"`
	Port      uint32                   `json:"port"`
	DebugPort uint32                   `json:"debug_port"`
	Chains    []string                 `json:"chains"` // TODO(olshansky): Integrate with utility to update this.
}

// TODO(olshansky): Add proper validator configuration validation.
func (v *Validator) Validate() error {
	return nil
}

// Mapping a validator from ID to the struct can make different parts of
// the consensus business logic easier & faster.
func ValidatorListToMap(validators []*Validator) (m ValMap) {
	m = make(ValMap, len(validators))
	for _, v := range validators {
		m[v.PublicKey.Address().String()] = v
	}
	return
}
