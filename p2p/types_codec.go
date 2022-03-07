package p2p

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sync"

	"github.com/pokt-network/pocket/p2p/types"
)

/*
 @ Implementation inspired from: perlin-network/noise
*/

type typesCodec struct {
	sync.RWMutex

	registered uint16 // 2bytes are important for encoding
	types      map[reflect.Type]uint16
	encoders   map[uint16]reflect.Value
	decoders   map[uint16]reflect.Value
}

var _ types.Codec = &typesCodec{}

func (c *typesCodec) Register(m interface{}, encoder, decoder interface{}) (uint16, error) {
	c.Lock()
	defer c.Unlock()

	t := reflect.TypeOf(m)
	e := reflect.ValueOf(encoder)
	d := reflect.ValueOf(decoder)

	if id, registered := c.types[t]; registered {
		return uint16(0), errors.New(fmt.Sprintf("typesCodec error: trying to register message of %+v which is already registered with id %d", t, id))
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
		return uint16(0), errors.New(fmt.Sprintf("typesCodec error: provided encoder for message type %+v is not valid, expected %s, but got %s", t, e, encoderSignature))
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
		return uint16(0), errors.New(fmt.Sprintf("typesCodec error: provided decoder for message type %+v is not valid, expected %s, but got %s", t, d, decoderSignature))
	}

	id := c.registered
	c.types[t] = id
	c.encoders[id] = e
	c.decoders[id] = d

	c.registered++

	return id, nil
}

func (c *typesCodec) Encode(msg interface{}) ([]byte, error) {
	c.Lock()
	defer c.Unlock()

	t := reflect.TypeOf(msg)

	if t.Kind() == reflect.Ptr {
		return nil, errors.New(fmt.Sprintf("typesCodec error: trying to encode a pointer"))
	}

	id, registered := c.types[t]
	if !registered {
		return nil, errors.New(fmt.Sprintf("typesCodec error: id not registered for message of type %+v", t))
	}

	encoder, registered := c.encoders[id]
	if !registered {
		return nil, errors.New(fmt.Sprintf("typesCodec error: encoder not registered for message of type %+v and id %d", t, id))
	}

	rt := encoder.Call([]reflect.Value{reflect.ValueOf(msg)})

	encoded := rt[0]
	erri := rt[1]
	if !erri.IsNil() {
		err := erri.Interface().(error)
		return nil, errors.New(fmt.Sprintf("typesCodec error: error encoding message of type %+v. err: %s", t, err.Error()))
	}

	if rt[0].IsNil() {
		return nil, errors.New(fmt.Sprintf("typesCodec error: encoded buffer is nil!"))
	}

	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, id)

	return append(buf, encoded.Interface().([]byte)...), nil
}

func (c *typesCodec) Decode(data []byte) (interface{}, error) {
	defer c.Unlock()
	c.Lock()

	if len(data) < 2 {
		return nil, errors.New(fmt.Sprintf("typesCodec error: cannot decode a byte with just message id and no actual message data. (data len %d)", len(data)))
	}

	id := binary.BigEndian.Uint16(data[:2])
	body := data[2:]

	decoder, registered := c.decoders[id]
	if !registered {
		return nil, errors.New(fmt.Sprintf("typesCodec error: no decoder is registered for message of id %d", id))
	}

	rt := decoder.Call([]reflect.Value{reflect.ValueOf(body)})
	msg := rt[0]
	erri := rt[1]

	if !erri.IsNil() {
		err := erri.Interface().(error)
		return nil, errors.New(fmt.Sprintf("typesCodec error: failed to decode message with id %d, err: %s", id, err.Error()))
	}

	return msg.Interface().(interface{}), nil
}

func NewTypesCodec() *typesCodec {
	return &typesCodec{
		registered: uint16(0),
		types:      make(map[reflect.Type]uint16, math.MaxUint16),
		encoders:   make(map[uint16]reflect.Value, math.MaxUint16),
		decoders:   make(map[uint16]reflect.Value, math.MaxUint16),
	}
}
