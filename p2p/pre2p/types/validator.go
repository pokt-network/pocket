package types

import pcrypto "github.com/pokt-network/pocket/shared/crypto"

type ValMap map[pcrypto.Address]*Validator

// TODO(olshansky): Review all the attributes of this struct.
type Validator struct {
	Address    string   `json:"address"`
	PublicKey  string   `json:"public_key"`
	PrivateKey string   `json:"private_key"`
	Jailed     bool     `json:"jailed"` // TODO(olshansky): Integrate with utility to update this.
	UPokt      uint64   `json:"upokt"`  // TODO(olshansky): Integrate with utility to update this.
	Host       string   `json:"host"`
	Port       uint32   `json:"port"`
	DebugPort  uint32   `json:"debug_port"`
	Chains     []string `json:"chains"` // TODO(olshansky): Integrate with utility to update this.
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
		m[v.A] = v
	}
	return
}
