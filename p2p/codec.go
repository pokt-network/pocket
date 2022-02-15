package p2p

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sync"
)

type Encoding string

const (
	Binary      Encoding = "bin"
	Utf8        Encoding = "utf8"
	Json        Encoding = "json"
	Grpc        Encoding = "grpc"
	Unsupported Encoding = "unsupported"
)

/*
 @
 @ Wire codec
 @
 @ Idea inspired by scuttlebutt's secure p2p wire protocol
*/
type wcodec struct {
	sync.RWMutex
}

/*
 @ * = 1 byte
 @ + = 1 bit
 @
 @ 9 bytes header
 @ -----header---][----body----]
 @ [*][****][****][****....****]
 @  0  1234  5678  ............
 @
 @ 0: flags [ + + + + + + + + ]
 @            <---> | | | <->
 @   empty:____|    | | |
 @                  | | |
 @ body wrapped?* __| | | encoding [0,0] = binary, [0,1] = utf8, [1,1] = json
 @                    | |
 @       is request __| |
 @ is erroror end ______|
 @
 @
 @ *: does the body have to be decoded at the application level (i.e by the domain codec)
 @
 @ 1234: request number as uint16, empty if not a request
 @ 5678: bodylength
 @
 @ body.
*/
func (c *wcodec) encode(encoding Encoding, iserror bool, reqnum uint32, data []byte, wrapped bool) []byte {
	c.Lock()
	defer c.Unlock()

	var flags byte = 0x00

	bodylength := make([]byte, 4)
	requestnumber := make([]byte, 4)

	binary.BigEndian.PutUint32(bodylength, uint32(len(data)))
	binary.BigEndian.PutUint32(requestnumber, uint32(reqnum))

	if wrapped {
		flags ^= 16 // set the fifth bit to 1
	}

	if reqnum != 0 {
		flags ^= 8 // setting the fourth bit to 1
	}

	if iserror {
		flags |= 4 // setting the third bit to 1
	}

	switch encoding {

	case Binary:
		// set the second and first bits to 0 (they are already at 0 from initialization)

	case Utf8:
		flags |= 1 // setting the first bit to 1, and the second to 0

	case Json:
		// set the first and second bit to 1
		flags |= 1
		flags |= 2
	}

	header := append([]byte{}, flags)
	header = append(header, requestnumber...)
	header = append(header, bodylength...)

	body := data

	payload := append(make([]byte, 0), header...)
	payload = append(payload, body...)
	return payload
}

func (c *wcodec) decode(wiredata []byte) (nonce uint32, enc Encoding, data []byte, wrapped bool, err error) {
	c.Lock()
	defer c.Unlock()

	header, body := wiredata[:9], wiredata[9:]
	flags := header[0]
	requestnum := header[1:5]
	bodylen := header[5:9]

	flagswitch, encoding, err := parseflag(flags)

	if err != nil {
		return 0, Unsupported, data, false, err
	}

	enc = encoding
	wrapped = flagswitch[4]
	isreq := flagswitch[3]
	if isreq {
		nonce = binary.BigEndian.Uint32(requestnum)
	} else {
		nonce = 0
	}

	iserr := flagswitch[2]
	if iserr {
		err = errors.New("")
	} else {
		err = nil
	}

	length := binary.BigEndian.Uint32(bodylen)
	data = body[:length]
	return
}

/*
 @
 @ Utils
 @
*/
func parseflag(f byte) (flagswitch []bool, e Encoding, err error) {
	if (f|31)^31 != 0 { // check if the first 3 bits are empty
		return nil, Unsupported, errors.New("codec wire flag error: invalid flag")
	}

	iswrapped := f & 16
	isreq := f & 8
	iserroreof := f & 4
	encoding := (f | 248) ^ 248

	flagswitch = make([]bool, 8)

	if uint(iswrapped) == 16 {
		flagswitch[4] = true
	} else {
		flagswitch[4] = false
	}

	if uint(isreq) == 8 {
		flagswitch[3] = true
	} else {
		flagswitch[3] = false
	}

	if uint(iserroreof) == 4 {
		flagswitch[2] = true
	} else {
		flagswitch[2] = false
	}

	switch uint(encoding) {
	case 0:
		e = Binary
	case 1:
		e = Utf8
	case 2:
		e = Json
	case 3:
		e = Grpc
	default:
		e = Unsupported
	}

	// filler values
	flagswitch[1] = false
	flagswitch[0] = false

	for i := 7; i > 4; i-- {
		flagswitch[i] = false
	}

	return flagswitch, e, nil
}

/*
 @
 @ Domain codec
 @
 @ Implementation inspired by perlin-network/noise
*/

type dcodec struct {
	sync.RWMutex

	registered uint16 // 2bytes are important for encoding
	types      map[reflect.Type]uint16
	encoders   map[uint16]reflect.Value
	decoders   map[uint16]reflect.Value
}

func (c *dcodec) register(m interface{}, encoder, decoder interface{}) (uint16, error) {
	c.Lock()
	defer c.Unlock()

	t := reflect.TypeOf(m)
	e := reflect.ValueOf(encoder)
	d := reflect.ValueOf(decoder)

	if id, registered := c.types[t]; registered {
		return uint16(0), errors.New(fmt.Sprintf("dcodec error: trying to register message of %+v which is already registered with id %d", t, id))
	}

	// expect decoders to be of type func(T) ([]byt, error)
	encoderSignature := reflect.FuncOf(
		[]reflect.Type{
			t,
		},
		[]reflect.Type{
			reflect.TypeOf(([]byte)(nil)),
			reflect.TypeOf((*error)(nil)).Elem(), // shenanigans to allow nil error
		},
		false, // not a variadic function
	)

	if e.Type() != encoderSignature {
		return uint16(0), errors.New(fmt.Sprintf("dcodec error: provided encoder for message type %+v is not valid, expected %s, but got %s", t, e, encoderSignature))
	}

	// expect decoders to be of type func([]byte) (T, error)
	decoderSignature := reflect.FuncOf(
		[]reflect.Type{
			reflect.TypeOf(([]byte)(nil)),
		},
		[]reflect.Type{
			t,                                    // T
			reflect.TypeOf((*error)(nil)).Elem(), // shenanigans to allow nil error
		},
		false, // not a variadic function
	)

	if d.Type() != decoderSignature {
		return uint16(0), errors.New(fmt.Sprintf("dcodec error: provided decoder for message type %+v is not valid, expected %s, but got %s", t, d, decoderSignature))
	}

	id := c.registered
	c.types[t] = id
	c.encoders[id] = e
	c.decoders[id] = d

	c.registered++

	return id, nil
}

func (c *dcodec) encode(msg interface{}) ([]byte, error) {
	c.Lock()
	defer c.Unlock()

	t := reflect.TypeOf(msg)

	if t.Kind() == reflect.Ptr {
		return nil, errors.New(fmt.Sprintf("dcodec error: trying to encode a pointer"))
	}

	id, registered := c.types[t]
	if !registered {
		return nil, errors.New(fmt.Sprintf("dcodec error: id not registered for message of type %+v", t))
	}

	encoder, registered := c.encoders[id]
	if !registered {
		return nil, errors.New(fmt.Sprintf("dcodec error: encoder not registered for message of type %+v and id %d", t, id))
	}

	rt := encoder.Call([]reflect.Value{reflect.ValueOf(msg)})

	encoded := rt[0]
	erri := rt[1]
	if !erri.IsNil() {
		err := erri.Interface().(error)
		return nil, errors.New(fmt.Sprintf("dcodec error: error encoding message of type %+v. err: %s", t, err.Error()))
	}

	if rt[0].IsNil() {
		return nil, errors.New(fmt.Sprintf("dcodec error: encoded buffer is nil!"))
	}

	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, id)

	return append(buf, encoded.Interface().([]byte)...), nil
}

func (c *dcodec) decode(data []byte) (interface{}, error) {
	defer c.Unlock()
	c.Lock()

	if len(data) < 2 {
		return nil, errors.New(fmt.Sprintf("dcodec error: cannot decode a byte with just message id and no actual message data. (data len %d)", len(data)))
	}

	id := binary.BigEndian.Uint16(data[:2])
	body := data[2:]

	decoder, registered := c.decoders[id]
	if !registered {
		return nil, errors.New(fmt.Sprintf("dcodec error: no decoder is registered for message of id %d", id))
	}

	rt := decoder.Call([]reflect.Value{reflect.ValueOf(body)})
	msg := rt[0]
	erri := rt[1]

	if !erri.IsNil() {
		err := erri.Interface().(error)
		return nil, errors.New(fmt.Sprintf("dcodec error: failed to decode message with id %d, err: %s", id, err.Error()))
	}

	return msg.Interface().(interface{}), nil
}

func NewDomainCodec() *dcodec {
	return &dcodec{
		registered: uint16(0),
		types:      make(map[reflect.Type]uint16, math.MaxUint16),
		encoders:   make(map[uint16]reflect.Value, math.MaxUint16),
		decoders:   make(map[uint16]reflect.Value, math.MaxUint16),
	}
}
