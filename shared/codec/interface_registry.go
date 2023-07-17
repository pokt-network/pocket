package codec

import (
	"errors"
	"fmt"
	reflect "reflect"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	// Exported accessor
	globalRegistry InterfaceRegistry = &registry{}

	// Errors
	ErrNotInterface               = errors.New("not an interface")
	ErrNotPointer                 = errors.New("not a pointer")
	ErrInterfaceAlreadyRegistered = errors.New("interface already registered")
	ErrNotRegistered              = errors.New("interface not registered")
	ErrCastingToProto             = errors.New("error casting to proto.Message")

	ErrDoesNotImplement = func(iface, impl string) error {
		return fmt.Errorf("%s does not implement %s", impl, iface)
	}
	ErrImplementationNotRegistered = func(impl, iface string) error {
		return fmt.Errorf("%s not registered as an implementation of %s", impl, iface)
	}
)

func init() {
	globalRegistry = newInterfaceRegistry()
}

type InterfaceRegistry interface {
	// RegisterInterface registers an interface with the registry.
	RegisterInterface(any) error
	// RegisterImplementations registers implementations of an interface with the registry.
	RegisterImplementations(iface any, impls ...proto.Message) error
	// MarshalInterface marshalls an interface into a byte array.
	MarshalInterface(any) ([]byte, error)
	// UnmarshalInterface unmarshalls a byte array into a generic interface
	UnmarshalInterface(anyPbBz []byte, iface any) error
}

type registry struct {
	// interfaces is a map of interfaces to their implementations.
	interfaces map[reflect.Type]map[string]reflect.Type
}

func newInterfaceRegistry() InterfaceRegistry {
	return &registry{
		interfaces: make(map[reflect.Type]map[string]reflect.Type),
	}
}

func GetInterfaceRegistry() InterfaceRegistry {
	return globalRegistry
}

func (r *registry) RegisterInterface(iface any) error {
	// get interface type
	ifaceType := reflect.TypeOf(iface).Elem()
	// check if it's an interface
	if ifaceType.Kind() != reflect.Interface {
		return ErrNotInterface
	}
	// check if its already registered
	if _, ok := r.interfaces[ifaceType]; ok {
		return ErrInterfaceAlreadyRegistered
	}
	// create a new map for implementations
	r.interfaces[ifaceType] = make(map[string]reflect.Type)
	return nil
}

func (r *registry) RegisterImplementations(iface any, impls ...proto.Message) error {
	// get interface type
	ifaceType := reflect.TypeOf(iface).Elem()
	// get map of implementations
	imap, ok := r.interfaces[ifaceType]
	if !ok {
		return ErrNotRegistered
	}
	// for each implementation provided register the proto name and type
	for _, impl := range impls {
		protoName := string(proto.MessageName(impl))
		implType := reflect.TypeOf(impl)
		if !implType.AssignableTo(ifaceType) {
			return ErrDoesNotImplement(ifaceType.Name(), protoName)
		}
		imap[protoName] = implType
	}
	return nil
}

func (r *registry) MarshalInterface(iface any) ([]byte, error) {
	// cast to a proto.Message
	protoIface, ok := iface.(proto.Message)
	if !ok {
		return nil, ErrCastingToProto
	}
	// convert to *anypb.Any{}
	anyIface, err := GetCodec().ToAny(protoIface)
	if err != nil {
		return nil, err
	}
	// marshal the *anypb.Any{}
	return GetCodec().Marshal(anyIface)
}

func (r *registry) UnmarshalInterface(anyPbBz []byte, iface any) error {
	// unmarshal into an *anypb.Any{}
	anyPb := new(anypb.Any)
	if err := GetCodec().Unmarshal(anyPbBz, anyPb); err != nil {
		return err
	}

	// get the value of the interface
	rv := reflect.ValueOf(iface)
	// ensure it is a pointer
	if rv.Kind() != reflect.Ptr {
		return ErrNotPointer
	}
	rt := rv.Elem().Type()
	// ensure it is an interface
	if rt.Kind() != reflect.Interface {
		return ErrNotInterface
	}

	// get the map of implementations
	imap, ok := r.interfaces[rt]
	if !ok {
		return ErrNotRegistered
	}
	// remove the anypb type prefix
	typeUrl := strings.TrimPrefix(anyPb.TypeUrl, "type.googleapis.com/")
	// get the implementation type
	implType, ok := imap[typeUrl]
	if !ok {
		return ErrImplementationNotRegistered(typeUrl, rt.Name())
	}

	// create a new instance of the implementation
	n := reflect.New(implType.Elem()).Interface().(proto.Message)
	// unmarshall the anypb value into the implementation
	if err := GetCodec().Unmarshal(anyPb.Value, n); err != nil {
		return err
	}

	// set the interface value to the implementation
	rv.Elem().Set(reflect.ValueOf(n))
	return nil
}
