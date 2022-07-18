package genesis

import (
	"encoding/hex"
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
)

type JsonHelper struct {
	Address   HexData `json:"address,omitempty"`
	PublicKey HexData `json:"public_key,omitempty"`
	Output    HexData `json:"output,omitempty"`
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

func (v *Validator) UnmarshalJSON(data []byte) error {
	var jh JsonHelper
	json.Unmarshal(data, &jh)

	protojson.Unmarshal(data, v)
	v.Address = jh.Address
	v.PublicKey = jh.PublicKey
	v.Output = jh.Output

	return nil
}
