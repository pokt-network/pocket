package types

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/modules"
)

type ord int

const (
	lt ord = iota - 1
	eq
	gt
)

func (h *Height) ToString() string {
	return fmt.Sprintf("%d-%d", h.RevisionNumber, h.RevisionHeight)
}

func (h *Height) IsZero() bool {
	return h.RevisionNumber == 0 && h.RevisionHeight == 0
}

func (h *Height) LT(other modules.Height) bool {
	return h.compare(other) == lt
}

func (h *Height) LTE(other modules.Height) bool {
	return h.compare(other) != gt
}

func (h *Height) GT(other modules.Height) bool {
	return h.compare(other) == gt
}

func (h *Height) GTE(other modules.Height) bool {
	return h.compare(other) != lt
}

func (h *Height) EQ(other modules.Height) bool {
	return h.compare(other) == eq
}

func (h *Height) Increment() modules.Height {
	return &Height{
		RevisionNumber: h.RevisionNumber,
		RevisionHeight: h.RevisionHeight + 1,
	}
}

func (h *Height) Decrement() modules.Height {
	return &Height{
		RevisionNumber: h.RevisionNumber,
		RevisionHeight: h.RevisionHeight - 1,
	}
}

func (h *Height) compare(other modules.Height) ord {
	if h.RevisionNumber > other.GetRevisionNumber() {
		return gt
	}
	if h.RevisionNumber < other.GetRevisionNumber() {
		return lt
	}
	if h.RevisionHeight > other.GetRevisionHeight() {
		return gt
	}
	if h.RevisionHeight < other.GetRevisionHeight() {
		return lt
	}
	return eq
}
