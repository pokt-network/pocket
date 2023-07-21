package types

import ics23 "github.com/cosmos/ics23/go"

var SmtSpec = &ProofSpec{
	LeafSpec: &LeafOp{
		Hash:         HashOp_SHA256,
		PrehashKey:   HashOp_SHA256,
		PrehashValue: HashOp_SHA256,
		Length:       LengthOp_NO_PREFIX,
		Prefix:       []byte{0},
	},
	InnerSpec: &InnerSpec{
		ChildOrder:      []int32{0, 1},
		ChildSize:       32,
		MinPrefixLength: 1,
		MaxPrefixLength: 1,
		EmptyChild:      make([]byte, 32),
		Hash:            HashOp_SHA256,
	},
	MaxDepth:                   256,
	PrehashKeyBeforeComparison: true,
}

func (p *ProofSpec) ConvertToIcs23ProofSpec() *ics23.ProofSpec {
	ics := new(ics23.ProofSpec)
	ics.LeafSpec = p.LeafSpec.ConvertToIcs23LeafOp()
	ics.InnerSpec = p.InnerSpec.ConvertToIcs23InnerSpec()
	ics.MaxDepth = p.MaxDepth
	ics.MinDepth = p.MinDepth
	ics.PrehashKeyBeforeComparison = p.PrehashKeyBeforeComparison
	return ics
}

func (l *LeafOp) ConvertToIcs23LeafOp() *ics23.LeafOp {
	ics := new(ics23.LeafOp)
	ics.Hash = l.Hash.ConvertToIcs23HashOp()
	ics.PrehashKey = l.PrehashKey.ConvertToIcs23HashOp()
	ics.PrehashValue = l.PrehashValue.ConvertToIcs23HashOp()
	ics.Length = l.Length.ConvertToIcs23LenthOp()
	ics.Prefix = l.Prefix
	return ics
}

func (i *InnerSpec) ConvertToIcs23InnerSpec() *ics23.InnerSpec {
	ics := new(ics23.InnerSpec)
	ics.ChildOrder = i.ChildOrder
	ics.ChildSize = i.ChildSize
	ics.MinPrefixLength = i.MinPrefixLength
	ics.MaxPrefixLength = i.MaxPrefixLength
	ics.EmptyChild = i.EmptyChild
	ics.Hash = i.Hash.ConvertToIcs23HashOp()
	return ics
}

func (h HashOp) ConvertToIcs23HashOp() ics23.HashOp {
	switch h {
	case HashOp_NO_HASH:
		return ics23.HashOp_NO_HASH
	case HashOp_SHA256:
		return ics23.HashOp_SHA256
	case HashOp_SHA512:
		return ics23.HashOp_SHA512
	case HashOp_KECCAK:
		return ics23.HashOp_KECCAK
	case HashOp_RIPEMD160:
		return ics23.HashOp_RIPEMD160
	case HashOp_BITCOIN:
		return ics23.HashOp_BITCOIN
	case HashOp_SHA512_256:
		return ics23.HashOp_SHA512_256
	default:
		panic("unknown hash op")
	}
}

func (l LengthOp) ConvertToIcs23LenthOp() ics23.LengthOp {
	switch l {
	case LengthOp_NO_PREFIX:
		return ics23.LengthOp_NO_PREFIX
	case LengthOp_VAR_PROTO:
		return ics23.LengthOp_VAR_PROTO
	case LengthOp_VAR_RLP:
		return ics23.LengthOp_VAR_RLP
	case LengthOp_FIXED32_BIG:
		return ics23.LengthOp_FIXED32_BIG
	case LengthOp_FIXED32_LITTLE:
		return ics23.LengthOp_FIXED32_LITTLE
	case LengthOp_FIXED64_BIG:
		return ics23.LengthOp_FIXED64_BIG
	case LengthOp_FIXED64_LITTLE:
		return ics23.LengthOp_FIXED64_LITTLE
	case LengthOp_REQUIRE_32_BYTES:
		return ics23.LengthOp_REQUIRE_32_BYTES
	case LengthOp_REQUIRE_64_BYTES:
		return ics23.LengthOp_REQUIRE_64_BYTES
	default:
		panic("unknown length op")
	}
}

func (i *InnerOp) ConvertToIcs23InnerOp() *ics23.InnerOp {
	ics := new(ics23.InnerOp)
	ics.Hash = i.Hash.ConvertToIcs23HashOp()
	ics.Prefix = i.Prefix
	ics.Suffix = i.Suffix
	return ics
}
