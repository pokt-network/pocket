package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (k KeybaseType) MarshalJSON() ([]byte, error) {
	return json.Marshal(KeybaseType_name[int32(k)])
}

// UnmarshalJSON converts the JSON string to a KeybaseType
func (k *KeybaseType) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	keybaseType, ok := KeybaseType_value[strings.ToUpper(value)]
	if !ok {
		return fmt.Errorf("invalid keybase type: %s", value)
	}
	*k = KeybaseType(keybaseType)
	return nil
}
