package types

type ord int

const (
	lt ord = iota
	eq
	gt
)

func (f *Fraction) LT(other *Fraction) bool {
	return f.compare(other) == lt
}

func (f *Fraction) GT(other *Fraction) bool {
	return f.compare(other) == gt
}

func (f *Fraction) EQ(other *Fraction) bool {
	return f.compare(other) == eq
}

func (f *Fraction) LTE(other *Fraction) bool {
	return f.compare(other) != gt
}

func (f *Fraction) GTE(other *Fraction) bool {
	return f.compare(other) != lt
}

func (f *Fraction) compare(other *Fraction) ord {
	comDenom := f.Denominator * other.Denominator
	aNum := f.Numerator * (comDenom / f.Denominator)
	bNum := other.Numerator * (comDenom / other.Denominator)
	if aNum < bNum {
		return lt
	}
	if aNum > bNum {
		return gt
	}
	return eq
}
