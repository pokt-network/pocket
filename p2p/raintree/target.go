package raintree

import (
	"fmt"
	"strings"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

type target struct {
	Address    cryptoPocket.Address
	ServiceUrl string

	Level                  uint32
	Percentage             float64
	AddrBookLengthAtHeight int
	Index                  int
	IsSelf                 bool
}

func (t target) DebugString(r *router) string {
	s := strings.Builder{}
	s.WriteString("[")
	serviceUrl := t.ServiceUrl
	if !t.IsSelf {
		s.WriteString(fmt.Sprintf(" (%s) ", serviceUrl))
	} else {
		s.WriteString(fmt.Sprintf("(self) %s ", serviceUrl))
	}

	for i := 1; i < t.AddrBookLengthAtHeight; i++ {
		serviceUrl := r.network.addrBookMap[r.network.addrList[i]].ServiceUrl
		if i == t.Index {
			s.WriteString(fmt.Sprintf(" **%s** ", serviceUrl))
		} else {
			s.WriteString(fmt.Sprintf(" %s ", serviceUrl))
		}
	}
	s.WriteString("]")
	return s.String()
}

func (t target) ShouldSendInternal() bool {
	return !t.IsSelf
}
