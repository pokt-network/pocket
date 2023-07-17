package codec

import (
	"fmt"
	reflect "reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type TestProtoInterface interface {
	proto.Message
	Stringify() string
}

func (t *TestProtoStructure) Stringify() string {
	return fmt.Sprintf("%d-%s-%v", t.Field1, t.Field2, t.Field3)
}

var _ TestProtoInterface = &TestProtoStructure{}

func TestInterfaceRegistry_RegisterInterface(t *testing.T) {
	registry := newInterfaceRegistry()

	// registers an interface
	err := registry.RegisterInterface((*TestProtoInterface)(nil))
	require.NoError(t, err)

	// fails to register the same interface
	err = registry.RegisterInterface((*TestProtoInterface)(nil))
	require.ErrorIs(t, err, ErrInterfaceAlreadyRegistered)

	// fails to register a non-interface
	tp := &TestProtoStructure{}
	err = registry.RegisterInterface(tp)
	require.ErrorIs(t, err, ErrNotInterface)
}

func TestInterfaceRegistry_RegisterImplementations(t *testing.T) {
	registry := newInterfaceRegistry()

	tp := &TestProtoStructure{
		Field1: 1,
		Field2: "test",
		Field3: true,
	}

	// fails to registers an implementation when interface isnt registered
	err := registry.RegisterImplementations(
		(*TestProtoInterface)(nil),
		tp,
	)
	require.ErrorIs(t, err, ErrNotRegistered)

	// registers an interface
	err = registry.RegisterInterface((*TestProtoInterface)(nil))
	require.NoError(t, err)

	// registers an implementation
	err = registry.RegisterImplementations(
		(*TestProtoInterface)(nil),
		tp,
	)
	require.NoError(t, err)

	// fails to register a message that doesn't implement the interface
	tp2 := &TestProtoStructure2{
		Field1: 1,
		Field2: "test",
	}
	err = registry.RegisterImplementations(
		(*TestProtoInterface)(nil),
		tp2,
	)
	require.Equal(t, err, ErrDoesNotImplement(
		reflect.TypeOf((*TestProtoInterface)(nil)).Elem().Name(),
		string(proto.MessageName(tp2)),
	))
}

func TestInterfaceRegistry_MarshalInterface(t *testing.T) {
	registry := newInterfaceRegistry()

	tp := &TestProtoStructure{
		Field1: 1,
		Field2: "test",
		Field3: true,
	}
	var pi TestProtoInterface = tp

	tp2 := &struct {
		Field1 int
		Field2 string
		Field3 bool
	}{
		Field1: 1,
		Field2: "test",
		Field3: true,
	}

	// fails to marhsal when interface isnt a proto.Message
	_, err := registry.MarshalInterface(tp2)
	require.ErrorIs(t, err, ErrCastingToProto)

	// marshals the interface
	bz, err := registry.MarshalInterface(pi)
	require.NoError(t, err)

	anyPb, err := GetCodec().ToAny(tp)
	require.NoError(t, err)
	pBz, err := GetCodec().Marshal(anyPb)
	require.NoError(t, err)

	require.Equal(t, pBz, bz)
}

func TestInterfaceRegistry_UnmarshalInterface(t *testing.T) {
	registry := newInterfaceRegistry()

	tp := &TestProtoStructure{
		Field1: 1,
		Field2: "test",
		Field3: true,
	}
	var pi TestProtoInterface = tp

	bz, err := registry.MarshalInterface(pi)
	require.NoError(t, err)

	// fails when iface is not a pointer
	var newPi TestProtoInterface
	err = registry.UnmarshalInterface(bz, newPi)
	require.ErrorIs(t, err, ErrNotPointer)

	// fails when iface is not an interface
	err = registry.UnmarshalInterface(bz, &tp)
	require.ErrorIs(t, err, ErrNotInterface)

	// fails to unmarshal when interface is not registered
	err = registry.UnmarshalInterface(bz, &newPi)
	require.ErrorIs(t, err, ErrNotRegistered)

	// registers an interface
	err = registry.RegisterInterface((*TestProtoInterface)(nil))
	require.NoError(t, err)

	// fails when implementation is not registered
	err = registry.UnmarshalInterface(bz, &newPi)
	require.Equal(t, err, ErrImplementationNotRegistered(
		string(proto.MessageName(tp)),
		reflect.TypeOf((*TestProtoInterface)(nil)).Elem().Name(),
	))

	// registers an implementation
	err = registry.RegisterImplementations(
		(*TestProtoInterface)(nil),
		tp,
	)
	require.NoError(t, err)

	// unmarshals the interface
	err = registry.UnmarshalInterface(bz, &newPi)
	require.NoError(t, err)
	require.Equal(t, pi.Stringify(), newPi.Stringify())
}
