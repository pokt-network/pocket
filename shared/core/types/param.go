package types

import (
	"fmt"
	"github.com/pokt-network/pocket/shared/crypto"
)

func (p Param) Hash() []byte {
	str := fmt.Sprintf("%s,%s,%d", p.Name, p.Value, p.Height)
	return crypto.SHA3Hash([]byte(str))
}

func (f Flag) Hash() []byte {
	str := fmt.Sprintf("%s,%s,%s,%d", f.Name, f.Value, f.Enabled, f.Height)
	return crypto.SHA3Hash([]byte(str))
}
