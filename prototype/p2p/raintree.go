package p2p

import (
	"math"
	"pocket/p2p/types"
)

/*
 @ Determine highest level possible in the tree (i.e number of layers)
*/
func getTopLevel(l *types.Peerlist) uint16 {
	fullListSize := float64(l.Size())

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
func pickLeft(srcid uint64, l *types.Peerlist) (lpos int) {
	lsize := float64(l.Size())

	ownposition := l.PositionOf(srcid)

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
func pickRight(srcid uint64, l *types.Peerlist) (rpos int) {
	lsize := float64(l.Size())

	ownposition := l.PositionOf(srcid)

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
func getTargetList(l *types.Peerlist, id uint64, topl, currl int) *types.Peerlist {
	tlsize := getTargetListSize(l.Size(), topl, currl)
	ownposition := l.PositionOf(id)

	slice := l.Slice()
	if ownposition+int(tlsize) > len(l.Slice()) {
		psublist := slice[ownposition:]
		csublist := slice[:int(tlsize)-len(psublist)]

		sublist := l.Copy()
		sublist.Update(psublist)

		return sublist.Concat(csublist)
	}

	sublist := l.Copy()
	slice = (&sublist).Slice()

	sublist.Update(slice[ownposition : ownposition+int(tlsize)])

	return &sublist
}

func rain(originatorId uint64, list *types.Peerlist, act func(id uint64, l, r *types.Peer, currentlevel int), root bool, fromlevel int) {
	var toplevel int = int(getTopLevel(list))
	var currentlevel int = toplevel

	if !root {
		currentlevel = fromlevel
	}

	for currentlevel > 0 {
		targetlist := getTargetList(list, originatorId, toplevel, currentlevel)

		var left, right *types.Peer
		{
			lpos := pickLeft(originatorId, targetlist)
			rpos := pickRight(originatorId, targetlist)

			left = targetlist.Get(lpos)
			right = targetlist.Get(rpos)
		}
		currentlevel--
		act(originatorId, left, right, currentlevel)
	}
}
