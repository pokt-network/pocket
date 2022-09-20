package genesis

// import (
// 	"encoding/hex"
// 	"encoding/json"

// 	"google.golang.org/protobuf/encoding/protojson"
// )

// // TODO_IN_THIS_COMMIT: See https://github.com/pokt-network/pocket/pull/139/files to remove this shit

// // HACK: Since the protocol actor protobufs (e.g. validator, fisherman, etc) use `bytes` for some
// // fields (e.g. `address`, `output`, `publicKey`), we need to use a helper struct to unmarshal the
// // the types when they are defined via json (e.g. genesis file, testing configurations, etc...).
// // Alternative solutions could include whole wrapper structs (i.e. duplication of schema definition),
// // using strings instead of bytes (i.e. major change with downstream effects) or avoid defining these
// // types in json altogether (i.e. limitation of usability).
// type JsonBytesLoaderHelper struct {
// 	Address   HexData `json:"address,omitempty"`
// 	PublicKey HexData `json:"public_key,omitempty"`
// 	Output    HexData `json:"output,omitempty"`
// }

// type HexData []byte

// func (h *HexData) UnmarshalJSON(data []byte) error {
// 	var s string
// 	if err := json.Unmarshal(data, &s); err != nil {
// 		return err
// 	}
// 	decoded, err := hex.DecodeString(s)
// 	if err != nil {
// 		return err
// 	}
// 	*h = HexData(decoded)
// 	return nil
// }

// func (v *Validator) UnmarshalJSON(data []byte) error {
// 	var jh JsonBytesLoaderHelper
// 	json.Unmarshal(data, &jh)

// 	protojson.Unmarshal(data, v)
// 	v.Address = jh.Address
// 	v.PublicKey = jh.PublicKey
// 	v.Output = jh.Output

// 	return nil
// }
