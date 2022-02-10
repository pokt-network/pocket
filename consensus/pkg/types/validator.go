package types

import "pocket/consensus/pkg/consensus/leader_election/vrf"

// A stable monotimally increasing integer used for
// identification, consensus, distributed key generation.
// Please note that this is needed and cannot be substituted
// for by address or public keys.
type NodeId uint32

type ValMap map[NodeId]*Validator

type Validator struct {
	NodeId    NodeId    `json:"node_id"`
	Address   string    `json:"address"`
	PublicKey PublicKey `json:"public_key"`
	Jailed    bool      `json:"jailed"` // TODO: Integrate with utility to update this.
	UPokt     uint64    `json:"upokt"`  // TODO: Integrate with utility to update this.
	Host      string    `json:"host"`
	Port      uint32    `json:"port"`
	DebugPort uint32    `json:"debug_port"`
	Chains    []string  `json:"chains"` // TODO: Integrate with utility to update this.

	// NOTE: Only having part of the attributes be supported by a config is bad practice.
	VRFVerificationKey vrf.VerificationKey // TODO: This will not be specified in any config file. Needs to be loaded from disk or retrieved from P2P network.
}

func (v *Validator) Validate() error {
	// log.Println("[TODO] Validator config validation is not implemented yet.")
	return nil
}

func ValidatorListToMap(validators []*Validator) (m ValMap) {
	m = make(ValMap, len(validators))
	for _, v := range validators {
		m[v.NodeId] = v
	}
	return
}
