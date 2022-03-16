package testutils

import "crypto/rand"

type DataChunk struct {
	Bytes   []byte // actual data
	Encoded []byte // data after encoding
	Error   error  // error encountered while reading or writing
	Length  uint   // the length written or read
}

func (d *DataChunk) Randomize(length int) {
	d.Bytes = GenerateByteLen(length)
	d.Length = uint(len(d.Bytes))
}

func NewDataChunk(l int) DataChunk {
	dchunk := DataChunk{
		Bytes:   make([]byte, 0),
		Encoded: make([]byte, 0),
		Length:  0,
		Error:   nil,
	}

	dchunk.Randomize(l)

	return dchunk
}

func GenerateByteLen(size int) []byte {
	buff := make([]byte, size)
	rand.Read(buff)
	return buff
}
