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

func (t target) DebugString(n *rainTreeNetwork) string {
	s := strings.Builder{}
	s.WriteString("[")
	serviceUrl := t.serviceUrl
	if !t.isSelf {
		fmt.Fprintf(&s, " (%s) ", serviceUrl)
	} else {
		fmt.Fprintf(&s, "(self) %s ", serviceUrl)
	}

	peersManagerStateView := n.peersManager.getNetworkView()
	for i := 1; i < t.addrBookLengthAtHeight; i++ {
		serviceUrl := peersManagerStateView.addrBookMap[peersManagerStateView.addrList[i]].ServiceUrl
		if i == t.index {
			fmt.Fprintf(&s, " **%s** ", serviceUrl)
		} else {
			fmt.Fprintf(&s, " %s ", serviceUrl)
		}
	}
	s.WriteString("]")
	return s.String()
}
