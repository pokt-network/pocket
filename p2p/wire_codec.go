package p2p

// TODO(derrandz): Deprecate this codec. Details in https://github.com/pokt-network/pocket/issues/57.

import (
	"encoding/binary"
	"errors"
	"sync"
)

type Encoding string

// TODO(derrandz): consider converting this to an enum.
const (
	Binary      Encoding = "bin"
	Utf8        Encoding = "utf8"
	Json        Encoding = "json"
	Grpc        Encoding = "grpc"
	Unsupported Encoding = "unsupported"
)

// Idea inspired by scuttlebutt's secure p2p wire protocol
//
// Layout of an encoded packet using wireCodec:
//
//    9 bytes header  a free length body not exceeding the max defined by configuration
//    [----header---][----body----]
//
//
// Breakdown of header:
//
// Symbolism used:____
//                    |
//    ---> * = 1 byte |
//    ---> + = 1 bit  |
//                    |
// --------------------
//
//    9 bytes header
//    [----header---][----body----]
//    [*][****][****][****....****]
//     0  1234  5678  ............
//
// Breakdown of the first byte reserved for flags:
//                  0 1 2 3 4 5 6 7
//  byte 0: flags [ + + + + + + + + ]
//                  <---> | | | <->
//         empty:____|    | | |  |
//                        | | |  |
//    is body wrapped?* __| | |  |---> wire encoding [0,0] = binary; [0,1] = utf8; [1,1] = json
//                          | |
//            is request? __| |
//      is erroror end? ______|
//
//
// *: does the body have to be decoded at the application level (i.e by the domain codec, think proto)
//
//  bytes 1234: request number/nonce/id as uint16, empty if not a request
//  bytes 5678: bodyLength
//

const (
	BodyLengthBytes    = 4
	RequestNonceLength = 4
)

const (
	WireCodecHeaderSize = 9
)

type wireCodec struct {
	sync.RWMutex
}

func (c *wireCodec) encode(isError bool, reqNonce uint32, data []byte, wrapped bool) []byte {
	c.Lock()
	defer c.Unlock()

	var flags byte = 0x00

	bodyLength := make([]byte, BodyLengthBytes)
	requestNonce := make([]byte, RequestNonceLength)

	binary.BigEndian.PutUint32(bodyLength, uint32(len(data)))
	binary.BigEndian.PutUint32(requestNonce, uint32(reqNonce))

	if wrapped {
		flags ^= 16 // set the fifth bit to 1
	}

	if reqNonce != 0 {
		flags ^= 8 // setting the fourth bit to 1
	}

	if isError {
		flags |= 4 // setting the third bit to 1
	}

	header := append([]byte{}, flags)
	header = append(header, requestNonce...)
	header = append(header, bodyLength...)

	body := data

	payload := append(make([]byte, 0), header...)
	payload = append(payload, body...)
	return payload
}

func (c *wireCodec) decode(wiredata []byte) (nonce uint32, data []byte, wrapped bool, err error) {
	c.Lock()
	defer c.Unlock()

	header, body := wiredata[:9], wiredata[9:]
	flags := header[0]
	requestNonce := header[1:5]
	bodylen := header[5:9]

	flagswitch, err := parseFlag(flags)

	if err != nil {
		return 0, data, false, err
	}

	wrapped = flagswitch[4]
	isReq := flagswitch[3]
	if isReq {
		nonce = binary.BigEndian.Uint32(requestNonce)
	} else {
		nonce = 0
	}

	isErr := flagswitch[2]
	if isErr {
		err = errors.New("")
	} else {
		err = nil
	}

	length := binary.BigEndian.Uint32(bodylen)
	data = body[:length]
	return
}

func (c *wireCodec) decodeHeader(header []byte) (flagswitch []bool, nonce uint32, isWrapped bool, bodyLength uint32, err error) {
	c.Lock()
	defer c.Unlock()

	flags := header[0]
	requestNonce := header[1:5]
	bodyLen := header[5:9]

	flagswitch, err = parseFlag(flags)

	if err != nil {
		return
	}

	isReq := flagswitch[3]
	if isReq {
		nonce = binary.BigEndian.Uint32(requestNonce)
	} else {
		nonce = 0
	}

	isWrapped = flagswitch[4]
	bodyLength = binary.BigEndian.Uint32(bodyLen)
	return
}

// Utility functions for the codec

// parseflag parses the first 1 byte of the header that constitutes the header flags.
// Flags are distributed on the 8 bits according to the codec's convention.
// Check the documentation at the top of the file to re-discover the flags represented on this 1 byte.
func parseFlag(f byte) (flagswitch []bool, err error) {
	if (f|31)^31 != 0 { // check if the first 3 bits are empty
		return nil, errors.New("codec wire flag error: invalid flag")
	}

	iswrapped := f & 16
	isReq := f & 8
	isErrOrEOF := f & 4

	flagswitch = make([]bool, 8)

	if uint(iswrapped) == 16 {
		flagswitch[4] = true
	} else {
		flagswitch[4] = false
	}

	if uint(isReq) == 8 {
		flagswitch[3] = true
	} else {
		flagswitch[3] = false
	}

	if uint(isErrOrEOF) == 4 {
		flagswitch[2] = true
	} else {
		flagswitch[2] = false
	}

	// filler values
	flagswitch[1] = false
	flagswitch[0] = false

	for i := 7; i > 4; i-- {
		flagswitch[i] = false
	}

	return flagswitch, nil
}

func NewWireCodec() *wireCodec {
	return &wireCodec{}
}
