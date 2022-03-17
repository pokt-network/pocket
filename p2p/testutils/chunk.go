package testutils

import "crypto/rand"

type DataChunk struct {
	Bytes   []byte // actual data
	Encoded []byte // data after encoding
	Error   error  // error encountered while reading or writing
	Length  uint   // the length written or read
}

func (d *DataChunk) Randomize(length int, encode func([]byte) []byte) {
	d.Bytes = randBytes(length)
	d.Encoded = encode(d.Bytes)
	d.Length = uint(len(d.Bytes))
}

func NewDataChunk(l int, encode func([]byte) []byte) DataChunk {
	dchunk := DataChunk{
		Bytes:   make([]byte, 0),
		Encoded: make([]byte, 0),
		Length:  0,
		Error:   nil,
	}

	dchunk.Randomize(l, encode)

	return dchunk
}

func randBytes(size int) []byte {
	buff := make([]byte, size)
	rand.Read(buff)
	return buff
}
