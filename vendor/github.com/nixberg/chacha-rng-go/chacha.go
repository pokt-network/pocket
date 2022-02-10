// Package chacha implements a ChaCha-based cryptographically secure
// pseudo-random number generator intended to be compatible with Rust crate
// rand_chacha.
package chacha

import (
	"crypto/rand"
	"encoding/binary"
	"math/bits"
)

type ChaCha struct {
	state        [16]uint32
	workingState [16]uint32
	rounds       int
	wordIndex    int
}

// New8 returns an instance of ChaCha8 seeded from a cryptographically secure
// random number generator.
func New8() (*ChaCha, error) {
	return new(8)
}

// New20 returns an instance of ChaCha20 seeded from a cryptographically secure
// random number generator.
func New20() (*ChaCha, error) {
	return new(20)
}

func new(rounds int) (*ChaCha, error) {
	var seed [8]uint32
	err := binary.Read(rand.Reader, binary.LittleEndian, &seed)
	if err != nil {
		return nil, err
	}
	return seeded(rounds, seed, 0), nil
}

// Zero8 returns an instance of ChaCha8 created from a seed of all zeros, set
// to the given stream.
func Zero8(stream uint64) *ChaCha {
	return seeded(8, [8]uint32{}, stream)
}

// Zero20 returns a instance of ChaCha20 created from a seed of all zeros, set
// to the given stream.
func Zero20(stream uint64) *ChaCha {
	return seeded(20, [8]uint32{}, stream)
}

// Seeded8 returns an instance of ChaCha8 created from the given seed, set to
// the given stream.
func Seeded8(seed [8]uint32, stream uint64) *ChaCha {
	return seeded(8, seed, stream)
}

// Seeded20 returns an instance of ChaCha20 created from the given seed, set to
// the given stream.
func Seeded20(seed [8]uint32, stream uint64) *ChaCha {
	return seeded(20, seed, stream)
}

func seeded(rounds int, seed [8]uint32, stream uint64) *ChaCha {
	return &ChaCha{
		state: [16]uint32{
			0x61707865,
			0x3320646e,
			0x79622d32,
			0x6b206574,
			seed[0],
			seed[1],
			seed[2],
			seed[3],
			seed[4],
			seed[5],
			seed[6],
			seed[7],
			0,
			0,
			uint32(stream),
			uint32(stream >> 32),
		},
		rounds:    rounds,
		wordIndex: 16,
	}
}

// Uint8 returns a pseudo-random 8-bit value as a uint8.
func (rng *ChaCha) Uint8() uint8 {
	return uint8(rng.Uint32())
}

// Uint16 returns a pseudo-random 16-bit value as a uint16.
func (rng *ChaCha) Uint16() uint16 {
	return uint16(rng.Uint32())
}

// Uint32 returns a pseudo-random 32-bit value as a uint32.
func (rng *ChaCha) Uint32() (result uint32) {
	if rng.wordIndex == 16 {
		rng.block()
		rng.incrementCounter()
		rng.wordIndex = 0
	}
	result = rng.workingState[rng.wordIndex]
	rng.wordIndex++
	return
}

// Uint64 returns a pseudo-random 64-bit value as a uint64.
func (rng *ChaCha) Uint64() uint64 {
	lo := uint64(rng.Uint32())
	hi := uint64(rng.Uint32())
	return (hi << 32) | lo
}

// Float32 returns, as a float32, a pseudo-random number in [0.0,1.0).
func (rng *ChaCha) Float32() float32 {
	return float32(rng.Uint32()>>8) * 0x1p-24
}

// Float64 returns, as a float64, a pseudo-random number in [0.0,1.0).
func (rng *ChaCha) Float64() float64 {
	return float64(rng.Uint64()>>11) * 0x1p-53
}

func (rng *ChaCha) block() {
	x0 := rng.state[0]
	x1 := rng.state[1]
	x2 := rng.state[2]
	x3 := rng.state[3]
	x4 := rng.state[4]
	x5 := rng.state[5]
	x6 := rng.state[6]
	x7 := rng.state[7]
	x8 := rng.state[8]
	x9 := rng.state[9]
	x10 := rng.state[10]
	x11 := rng.state[11]
	x12 := rng.state[12]
	x13 := rng.state[13]
	x14 := rng.state[14]
	x15 := rng.state[15]

	for i := 0; i < rng.rounds; i += 2 {
		x0, x4, x8, x12 = quarterRound(x0, x4, x8, x12)
		x1, x5, x9, x13 = quarterRound(x1, x5, x9, x13)
		x2, x6, x10, x14 = quarterRound(x2, x6, x10, x14)
		x3, x7, x11, x15 = quarterRound(x3, x7, x11, x15)

		x0, x5, x10, x15 = quarterRound(x0, x5, x10, x15)
		x1, x6, x11, x12 = quarterRound(x1, x6, x11, x12)
		x2, x7, x8, x13 = quarterRound(x2, x7, x8, x13)
		x3, x4, x9, x14 = quarterRound(x3, x4, x9, x14)
	}

	rng.workingState[0] = x0 + rng.state[0]
	rng.workingState[1] = x1 + rng.state[1]
	rng.workingState[2] = x2 + rng.state[2]
	rng.workingState[3] = x3 + rng.state[3]
	rng.workingState[4] = x4 + rng.state[4]
	rng.workingState[5] = x5 + rng.state[5]
	rng.workingState[6] = x6 + rng.state[6]
	rng.workingState[7] = x7 + rng.state[7]
	rng.workingState[8] = x8 + rng.state[8]
	rng.workingState[9] = x9 + rng.state[9]
	rng.workingState[10] = x10 + rng.state[10]
	rng.workingState[11] = x11 + rng.state[11]
	rng.workingState[12] = x12 + rng.state[12]
	rng.workingState[13] = x13 + rng.state[13]
	rng.workingState[14] = x14 + rng.state[14]
	rng.workingState[15] = x15 + rng.state[15]
}

func (rng *ChaCha) incrementCounter() {
	rng.state[12]++
	if rng.state[12] == 0 {
		rng.state[13]++
		if rng.state[13] == 0 {
			panic("chacha: counter overflow")
		}
	}
}

func quarterRound(a, b, c, d uint32) (uint32, uint32, uint32, uint32) {
	a += b
	d = bits.RotateLeft32(d^a, 16)

	c += d
	b = bits.RotateLeft32(b^c, 12)

	a += b
	d = bits.RotateLeft32(d^a, 8)

	c += d
	b = bits.RotateLeft32(b^c, 7)

	return a, b, c, d
}
