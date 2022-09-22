package raintree

import (
	"fmt"
	"strings"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

type target struct {
	address    cryptoPocket.Address
	serviceUrl string

	level                  uint32
	percentage             float64
	addrBookLengthAtHeight int
	index                  int
	isSelf                 bool
}

func (t target) DebugString(r *router) string {
	s := strings.Builder{}
	s.WriteString("[")
	serviceUrl := t.serviceUrl
	if !t.isSelf {
		fmt.Fprintf(&s, " (%s) ", serviceUrl)
	} else {
		fmt.Fprintf(&s, "(self) %s ", serviceUrl)
	}

	for i := 1; i < t.addrBookLengthAtHeight; i++ {
		serviceUrl := r.network.addrBookMap[r.network.addrList[i]].ServiceUrl
		if i == t.index {
			fmt.Fprintf(&s, " **%s** ", serviceUrl)
		} else {
			fmt.Fprintf(&s, " %s ", serviceUrl)
		}
	}
	s.WriteString("]")
	return s.String()
}

func (t target) ShouldSendInternal() bool {
	return !t.isSelf
}
