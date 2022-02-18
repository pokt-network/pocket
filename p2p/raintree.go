package p2p

import (
	"math"
)

/*
 @ Determine highest level possible in the tree (i.e number of layers)
*/
func getTopLevel(list *plist) uint16 {
	fullListSize := float64(len(list.elements))

	return uint16(
		math.Ceil(
			math.Round(
				(math.Log(fullListSize)/math.Log(3.0))*100,
			)/100,
		),
	) + 1
}

/*
 @ Determine target list size based on full list
*/
func getTargetListSize(fullListSize, topl, currl int) float64 {
	tlsize := math.Round(float64(fullListSize) * math.Pow(float64(0.66), float64(topl-currl)))
	return tlsize
}

/*
 @ Pick left branch of the tree in the target list
*/
func pickLeft(srcid uint64, l *plist) (lpos int) {
	lsize := float64(l.size())

	ownposition := l.positionof(srcid)

	lpos = int(math.Round(float64(ownposition)+lsize/float64(1.5))) + 1
	lpos = int(lsize) - lpos + 1

	if lpos > int(lsize) {
		lpos -= int(lsize) // rollover
	}

	return
}

/*
 @ Pick right branch of the tree in the target list
*/
func pickRight(srcid uint64, l *plist) (rpos int) {
	lsize := float64(l.size())

	ownposition := l.positionof(srcid)

	rpos = int(math.Round(float64(ownposition)+lsize/float64(3))) + 1
	rpos = int(lsize) - rpos + 1

	if rpos > int(lsize) { // rollover if needed // rollover if needed // rollover if needed // rollover if needed
		rpos -= int(lsize)
	}

	return
}

/*
 @ Retrieve the target list out of the full list
*/
func getTargetList(l *plist, id uint64, topl, currl int) *plist {
	tlsize := getTargetListSize(l.size(), topl, currl)
	ownposition := l.positionof(id)

	slice := l.slice()
	if ownposition+int(tlsize) > len(l.slice()) {
		psublist := slice[ownposition:]
		csublist := slice[:int(tlsize)-len(psublist)]

		sublist := l.copy()
		sublist.update(psublist)

		return sublist.concat(csublist)
	}

	sublist := l.copy()
	slice = (&sublist).slice()

	sublist.update(slice[ownposition : ownposition+int(tlsize)])

	return &sublist
}
