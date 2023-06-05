package modules

// ModuleFactoryWithOptions implements a `#Create()` factory method which takes
// a variadic `ModuleOption` argument(s) and returns a `Module`and an error.
type ModuleFactoryWithOptions FactoryWithOptions[Module, ModuleOption]

// Factory implements a `#Create()` factory method which takes a bus and returns
// a value of type T and an error.
type Factory[T interface{}] interface {
	Create(bus Bus) (T, error)
}

// FactoryWithConfig implements a `#Create()` factory method which takes a bus and
// a required "config" argument of type K and returns a value of type T and an error.
// TECHDEBT: apply enforcement across applicable "sub-modules" (see: `p2p/raintree/router.go`: `raintTreeFactory`)
type FactoryWithConfig[T interface{}, K interface{}] interface {
	Create(bus Bus, cfg K) (T, error)
}

// FactoryWithOptions implements a `#Create()` factory method which takes a
// variadic "optional" argument(s) of type O and returns a value of type T
// and an error.
// TECHDEBT: apply enforcement across applicable "sub-modules"
type FactoryWithOptions[T interface{}, O interface{}] interface {
	Create(bus Bus, opts ...O) (T, error)
}

// FactoryWithConfigAndOptions implements a `#Create()` factory method which
// takes both a required "config" argument of type K and a variadic "optional"
// argument(s) of type O and returns a value of type T and an error.
// TECHDEBT: apply enforcement across applicable "sub-modules"
type FactoryWithConfigAndOptions[T interface{}, K interface{}, O interface{}] interface {
	Create(bus Bus, cfg K, opts ...O) (T, error)
}
