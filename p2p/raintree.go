package p2p

import (
	"sort"

	"math"
)

type RainTree interface {
	SetLeafs([]peerInfo)
	SetRoot(int)
	Traverse(bool, int, func(int, peerInfo, peerInfo, int) error) error
	GetTargetList(int, int) []int
	GetTopLevel() int
	GetTargetListSize(int, int) float64
	PickLeft([]int) int
	PickRight([]int) int
	PositionOf(int) int
	CopySortedList() []int
	GetByPosition(int) peerInfo
	GetSortedList() []int
}

type rainTree struct {
	list       []peerInfo
	sortedList []int
	root       int
}

func NewRainTree() RainTree {
	return &rainTree{}

}

func (rt *rainTree) SetLeafs(leafs []peerInfo) {
	rt.list = leafs

	for _, k := range leafs {
		rt.sortedList = append(rt.sortedList, k.ID)
	}

	sortedListSlice := rt.sortedList[:]
	sort.Ints(sortedListSlice)
}

func (rt *rainTree) SetRoot(id int) {
	rt.root = id
}

// Determine highest level possible in the tree (i.e number of layers)
func (rt *rainTree) GetTopLevel() int {
	fullListSize := float64(len(rt.sortedList))

	return int(
		math.Ceil(
			math.Round(
				(math.Log(fullListSize)/math.Log(3.0))*100,
			)/100,
		),
	) + 1
}

// Determine target list size based on full list
func (rt *rainTree) GetTargetListSize(topLevel, currentLevel int) float64 {
	tlsize := math.Round(float64(len(rt.sortedList)) * math.Pow(float64(0.66), float64(topLevel-currentLevel)))
	return tlsize
}

// Pick left branch of the tree in the target list
func (rt *rainTree) PickLeft(targetList []int) (lpos int) {
	lsize := float64(len(targetList))
	ownposition := rt.PositionOf(rt.root)

	lpos = int(math.Round(float64(ownposition)+lsize/float64(1.5))) + 1

	if lpos > int(lsize) {
		lpos -= int(lsize) // rollover
	} else {
		lpos = int(lsize) - lpos + 1
	}

	return
}

// Pick right branch of the tree in the target list
func (rt *rainTree) PickRight(targetList []int) (rpos int) {
	lsize := float64(len(targetList))

	ownposition := rt.PositionOf(rt.root)

	rpos = int(math.Round(float64(ownposition)+lsize/float64(3))) + 1

	if rpos > int(lsize) { // rollover if needed
		rpos -= int(lsize)
	} else {
		rpos = int(lsize) - rpos + 1
	}

	return
}

// Retrieve the target list out of the full list
func (rt *rainTree) GetTargetList(topl, currl int) []int {
	tlsize := rt.GetTargetListSize(topl, currl)
	ownposition := rt.PositionOf(rt.root)

	list := rt.CopySortedList()
	if ownposition+int(tlsize) > len(list) { // rolling over logic
		psublist := list[ownposition:]
		csublist := list[:int(tlsize)-len(psublist)]

		return append(psublist, csublist...)
	}

	return list[ownposition : ownposition+int(tlsize)]
}

func (rt *rainTree) Traverse(root bool, fromlevel int, act func(origin int, l, r peerInfo, currentlevel int) error) error {
	var toplevel int = int(rt.GetTopLevel())
	var currentlevel int = toplevel

	if !root {
		currentlevel = fromlevel
	}

	for currentlevel > 0 {
		targetList := rt.GetTargetList(toplevel, currentlevel)

		var left, right peerInfo
		lpos := rt.PickLeft(targetList)
		rpos := rt.PickRight(targetList)

		left = rt.GetByPosition(lpos)
		right = rt.GetByPosition(rpos)

		currentlevel--
		if err := act(rt.root, left, right, currentlevel); err != nil {
			return err
		}
	}

	return nil
}

func (rt *rainTree) CopySortedList() []int {
	return append(make([]int, 0), rt.sortedList...)
}

func (rt *rainTree) GetByPosition(pos int) peerInfo {
	position := -1
	id := rt.sortedList[pos]
	for i, v := range rt.list {
		if v.ID == id {
			position = i
		}
	}
	return rt.list[position]
}

func (rt *rainTree) GetSortedList() []int {
	return rt.sortedList
}

func (rt *rainTree) PositionOf(id int) int {
	position := -1

	for i, v := range rt.sortedList {
		if v == id {
			position = i
			break
		}
	}

	return position
}
