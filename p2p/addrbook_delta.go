package p2p

import typesP2P "github.com/pokt-network/pocket/p2p/types"

// getAddrBookDelta returns the difference between two AddrBook slices
func getAddrBookDelta(before, after typesP2P.AddrBook) (added, removed []*typesP2P.NetworkPeer) {
	oldMap := make(map[string]*typesP2P.NetworkPeer)
	for _, np := range before {
		oldMap[np.Address.String()] = np
	}

	for _, np := range after {
		if _, ok := oldMap[np.Address.String()]; !ok {
			added = append(added, np)
			continue
		}
		delete(oldMap, np.Address.String())
	}

	for _, u := range oldMap {
		removed = append(removed, u)
	}

	return
}
