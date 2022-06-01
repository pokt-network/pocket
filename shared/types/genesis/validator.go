package genesis

import (
	"encoding/hex"
	"encoding/json"
)

// TECHDEBT(olshansky): This is a wrapper around the generated `Validator.go`
// because we cannot load `[]byte` from JSON.
type ValidatorJsonCompatibleWrapper struct {
	Address         HexData `json:"address,omitempty"`
	PublicKey       HexData `json:"public_key,omitempty"`
	Paused          bool    `json:"paused,omitempty"`
	Status          int32   `json:"status,omitempty"`
	ServiceUrl      string  `json:"service_url,omitempty"`
	StakedTokens    string  `json:"staked_tokens,omitempty"`
	MissedBlocks    uint32  `json:"missed_blocks,omitempty"`
	PausedHeight    uint64  `json:"paused_height,omitempty"`
	UnstakingHeight int64   `json:"unstaking_height,omitempty"`
	Output          HexData `json:"output,omitempty"`
}

type HexData []byte

func (h *HexData) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	*h = HexData(decoded)
	return nil
}

func (v *ValidatorJsonCompatibleWrapper) ValidateBasic() error {
	return nil
}

func (v *ValidatorJsonCompatibleWrapper) Validator() *Validator {
	return &Validator{
		Address:         v.Address,
		PublicKey:       v.PublicKey,
		Paused:          v.Paused,
		Status:          v.Status,
		ServiceUrl:      v.ServiceUrl,
		StakedTokens:    v.StakedTokens,
		MissedBlocks:    v.MissedBlocks,
		PausedHeight:    v.PausedHeight,
		UnstakingHeight: v.UnstakingHeight,
		Output:          v.Output,
	}
}

func GetValidators(vals []*ValidatorJsonCompatibleWrapper) (validators []*Validator) {
	validators = make([]*Validator, len(vals))
	for i, v := range vals {
		validators[i] = &Validator{
			Address:         v.Address,
			PublicKey:       v.PublicKey,
			Paused:          v.Paused,
			Status:          v.Status,
			ServiceUrl:      v.ServiceUrl,
			StakedTokens:    v.StakedTokens,
			MissedBlocks:    v.MissedBlocks,
			PausedHeight:    v.PausedHeight,
			UnstakingHeight: v.UnstakingHeight,
			Output:          v.Output,
		}
	}
	return
}
