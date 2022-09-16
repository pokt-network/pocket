package types

type AddrList []string

func (ab *AddrList) Find(address string) (index int, found bool) {
	if ab == nil {
		return 0, false
	}
	addressBook := *ab
	for i, a := range addressBook {
		if a == address {
			return i, true
		}
	}
	return 0, false
}
