package types

import (
	"fmt"
	"sort"
)

type AddrBook []*NetworkPeer

func (ab *AddrBook) ToListAndMap(selfAddressString string) (addrList AddrList, addrBookMap AddrBookMap, err error) {
	addrBook := *ab
	// OPTIMIZE(olshansky): This is a very naive approach for now that recomputes everything every time that we can optimize later
	addrBookMap = make(map[string]*NetworkPeer, len(addrBook))
	addrList = make([]string, len(addrBook))
	for i, peer := range addrBook {
		addr := peer.Address.String()
		addrList[i] = addr
		addrBookMap[addr] = peer
	}
	sort.Strings(addrList)
	if i, ok := addrList.Find(selfAddressString); ok {
		// The list is sorted lexicographically above, but is reformatted below so this addr of this node
		// is always the first in the list. This makes RainTree propagation easier to compute and interpret.
		addrList = append(addrList[i:], addrList[0:i]...)
	} else {
		return nil, nil, fmt.Errorf("self address not found for %s in addrBook so this client can send messages but does not propagate them", selfAddressString)
	}
	return
}
