package raintree

import (
	"math"
	"sort"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// Refer to the specification for a formal description and proof of how the constants
// and math functions here were determined.
const (
	firstMsgTargetPercentage  = float64(1) / float64(3)
	secondMsgTargetPercentage = float64(2) / float64(3)
	shrinkagePercentage       = float64(2) / float64(3)
	maxLevelsLogBase          = float64(3)
)

// Whenever `addrBook` changes, we also need to update `addrBookMap` and `addrList`.
// TODO(olshansky): This is a very naive approach for now that recomputes everything every time that we can optimize later
func (n *rainTreeNetwork) handleAddrBookUpdates() error {
	n.addrBookMap = make(map[string]*typesPre2P.NetworkPeer, len(n.addrBook))
	n.addrList = make([]string, len(n.addrBook))
	for i, peer := range n.addrBook {
		addr := peer.Address.String()
		n.addrList[i] = addr
		n.addrBookMap[addr] = peer
	}
	n.maxNumLevels = n.getMaxAddrBookLevels()

	sort.Strings(n.addrList)
	i := n.getSelfIndexInAddrBook()
	n.addrList = append(n.addrList[i:len(n.addrList)], n.addrList[0:i]...)

	return nil
}

func (n *rainTreeNetwork) getFirstTargetAddr() (cryptoPocket.Address, bool) {
	i := int(firstMsgTargetPercentage * float64(len(n.addrList)))
	addrStr := n.addrList[i]
	return n.addrBookMap[addrStr].Address, true
}

func (n *rainTreeNetwork) getSecondTargetAddr() (cryptoPocket.Address, bool) {
	i := int(secondMsgTargetPercentage * float64(len(n.addrList)))
	addrStr := n.addrList[i]
	return n.addrBookMap[addrStr].Address, true
}

func (n *rainTreeNetwork) getSelfIndexInAddrBook() int {
	addrString := n.addr.String()
	for i, addr := range n.addrList {
		if addr == addrString {
			return i
		}
	}
	return -1
}

func (n *rainTreeNetwork) getMaxAddrBookLevels() uint32 {
	addrBookSize := float64(len(n.addrBook))
	return uint32(math.Ceil(log(addrBookSize)*100) / 100)
}

func log(x float64) float64 {
	return math.Log(x) / math.Log(maxLevelsLogBase)
}
