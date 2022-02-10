package chacha

// FillUint8 fills the given slice with pseudo-random uint8s.
func (rng *ChaCha) FillUint8(slice []uint8) {
	count := len(slice)
	tailCount := len(slice) % 4

	for i := 0; i < (count - tailCount); i += 4 {
		word := rng.Uint32()
		slice[i+0] = uint8(word)
		slice[i+1] = uint8(word >> 8)
		slice[i+2] = uint8(word >> 16)
		slice[i+3] = uint8(word >> 24)
	}

	if tailCount > 0 {
		word := rng.Uint32()
		for i := tailCount; i > 0; i-- {
			slice[count-i] = uint8(word)
			word >>= 8
		}
	}
}

// FillUint16 fills the given slice with pseudo-random uint16s.
func (rng *ChaCha) FillUint16(slice []uint16) {
	count := len(slice)
	tailCount := len(slice) % 2

	for i := 0; i < (count - tailCount); i += 2 {
		word := rng.Uint32()
		slice[i+0] = uint16(word)
		slice[i+1] = uint16(word >> 16)
	}

	if tailCount > 0 {
		slice[count-1] = uint16(rng.Uint32())
	}
}

// FillUint32 fills the given slice with pseudo-random uint32s.
func (rng *ChaCha) FillUint32(slice []uint32) {
	for i := range slice {
		slice[i] = rng.Uint32()
	}
}

// FillUint64 fills the given slice with pseudo-random uint64s.
func (rng *ChaCha) FillUint64(slice []uint64) {
	for i := range slice {
		slice[i] = rng.Uint64()
	}
}

// FillFloat32 fills the given slice with pseudo-random float32s.
func (rng *ChaCha) FillFloat32(slice []float32) {
	for i := range slice {
		slice[i] = rng.Float32()
	}
}

// FillFloat64 fills the given slice with pseudo-random float64s.
func (rng *ChaCha) FillFloat64(slice []float64) {
	for i := range slice {
		slice[i] = rng.Float64()
	}
}
