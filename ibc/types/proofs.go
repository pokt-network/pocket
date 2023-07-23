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
	if p == nil {
		return nil
	}
	ics := new(ics23.ProofSpec)
	ics.LeafSpec = p.LeafSpec.convertToIcs23LeafOp()
	ics.InnerSpec = p.InnerSpec.convertToIcs23InnerSpec()
	ics.MaxDepth = p.MaxDepth
	ics.MinDepth = p.MinDepth
	ics.PrehashKeyBeforeComparison = p.PrehashKeyBeforeComparison
	return ics
}

func ConvertFromIcs23ProofSpec(p *ics23.ProofSpec) *ProofSpec {
	if p == nil {
		return nil
	}
	spc := new(ProofSpec)
	spc.LeafSpec = convertFromIcs23LeafOp(p.LeafSpec)
	spc.InnerSpec = convertFromIcs23InnerSpec(p.InnerSpec)
	spc.MaxDepth = p.MaxDepth
	spc.MinDepth = p.MinDepth
	spc.PrehashKeyBeforeComparison = p.PrehashKeyBeforeComparison
	return spc
}

func (l *LeafOp) convertToIcs23LeafOp() *ics23.LeafOp {
	if l == nil {
		return nil
	}
	ics := new(ics23.LeafOp)
	ics.Hash = l.Hash.convertToIcs23HashOp()
	ics.PrehashKey = l.PrehashKey.convertToIcs23HashOp()
	ics.PrehashValue = l.PrehashValue.convertToIcs23HashOp()
	ics.Length = l.Length.convertToIcs23LenthOp()
	ics.Prefix = l.Prefix
	return ics
}

func convertFromIcs23LeafOp(l *ics23.LeafOp) *LeafOp {
	if l == nil {
		return nil
	}
	op := new(LeafOp)
	op.Hash = convertFromIcs23HashOp(l.Hash)
	op.PrehashKey = convertFromIcs23HashOp(l.PrehashKey)
	op.PrehashValue = convertFromIcs23HashOp(l.PrehashValue)
	op.Length = convertFromIcs23LengthOp(l.Length)
	op.Prefix = l.Prefix
	return op
}

func (i *InnerSpec) convertToIcs23InnerSpec() *ics23.InnerSpec {
	if i == nil {
		return nil
	}
	ics := new(ics23.InnerSpec)
	ics.ChildOrder = i.ChildOrder
	ics.ChildSize = i.ChildSize
	ics.MinPrefixLength = i.MinPrefixLength
	ics.MaxPrefixLength = i.MaxPrefixLength
	ics.EmptyChild = i.EmptyChild
	ics.Hash = i.Hash.convertToIcs23HashOp()
	return ics
}

func convertFromIcs23InnerSpec(i *ics23.InnerSpec) *InnerSpec {
	if i == nil {
		return nil
	}
	spec := new(InnerSpec)
	spec.ChildOrder = i.ChildOrder
	spec.ChildSize = i.ChildSize
	spec.MinPrefixLength = i.MinPrefixLength
	spec.MaxPrefixLength = i.MaxPrefixLength
	spec.EmptyChild = i.EmptyChild
	spec.Hash = convertFromIcs23HashOp(i.Hash)
	return spec
}

func (h HashOp) convertToIcs23HashOp() ics23.HashOp {
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

func convertFromIcs23HashOp(h ics23.HashOp) HashOp {
	switch h {
	case ics23.HashOp_NO_HASH:
		return HashOp_NO_HASH
	case ics23.HashOp_SHA256:
		return HashOp_SHA256
	case ics23.HashOp_SHA512:
		return HashOp_SHA512
	case ics23.HashOp_KECCAK:
		return HashOp_KECCAK
	case ics23.HashOp_RIPEMD160:
		return HashOp_RIPEMD160
	case ics23.HashOp_BITCOIN:
		return HashOp_BITCOIN
	case ics23.HashOp_SHA512_256:
		return HashOp_SHA512_256
	default:
		panic("unknown hash op")
	}
}

func (l LengthOp) convertToIcs23LenthOp() ics23.LengthOp {
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

func convertFromIcs23LengthOp(l ics23.LengthOp) LengthOp {
	switch l {
	case ics23.LengthOp_NO_PREFIX:
		return LengthOp_NO_PREFIX
	case ics23.LengthOp_VAR_PROTO:
		return LengthOp_VAR_PROTO
	case ics23.LengthOp_VAR_RLP:
		return LengthOp_VAR_RLP
	case ics23.LengthOp_FIXED32_BIG:
		return LengthOp_FIXED32_BIG
	case ics23.LengthOp_FIXED32_LITTLE:
		return LengthOp_FIXED32_LITTLE
	case ics23.LengthOp_FIXED64_BIG:
		return LengthOp_FIXED64_BIG
	case ics23.LengthOp_FIXED64_LITTLE:
		return LengthOp_FIXED64_LITTLE
	case ics23.LengthOp_REQUIRE_32_BYTES:
		return LengthOp_REQUIRE_32_BYTES
	case ics23.LengthOp_REQUIRE_64_BYTES:
		return LengthOp_REQUIRE_64_BYTES
	default:
		panic("unknown length op")
	}
}

func (i *InnerOp) convertToIcs23InnerOp() *ics23.InnerOp {
	if i == nil {
		return nil
	}
	ics := new(ics23.InnerOp)
	ics.Hash = i.Hash.convertToIcs23HashOp()
	ics.Prefix = i.Prefix
	ics.Suffix = i.Suffix
	return ics
}

func convertFromIcs23InnerOp(i *ics23.InnerOp) *InnerOp {
	if i == nil {
		return nil
	}
	op := new(InnerOp)
	op.Hash = convertFromIcs23HashOp(i.Hash)
	op.Prefix = i.Prefix
	op.Suffix = i.Suffix
	return op
}
