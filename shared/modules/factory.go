package modules

// ModuleFactoryWithOptions implements a `#Create()` factory method which takes
// a variadic `ModuleOption` argument(s) and returns a `Module`and an error.
type ModuleFactoryWithOptions FactoryWithOptions[Module, ModuleOption]

// Factory implements a `#Create()` factory method which takes a bus and returns
// a value of type M and an error.
type Factory[M any] interface {
	Create(bus Bus) (M, error)
}

// FactoryWithConfig implements a `#Create()` factory method which takes a bus and
// a required "config" argument of type C and returns a value of type M and an error.
// TECHDEBT: apply enforcement across applicable "sub-modules" (see: `p2p/raintree/router.go`: `raintTreeFactory`)
type FactoryWithConfig[T any, C any] interface {
	Create(bus Bus, cfg C) (T, error)
}

// FactoryWithOptions implements a `#Create()` factory method which takes a bus
// and a variadic "optional" argument(s) of type O and returns a value of type M
// and an error.
// TECHDEBT: apply enforcement across applicable "sub-modules"
type FactoryWithOptions[M any, O any] interface {
	Create(bus Bus, opts ...O) (M, error)
}

// FactoryWithConfigAndOptions implements a `#Create()` factory method which takes
// a bus, a required "config" argument of type C, and a variadic (optional)
// argument(s) of type O and returns a value of type M and an error.
// TECHDEBT: apply enforcement across applicable "sub-modules"
type FactoryWithConfigAndOptions[M any, C any, O any] interface {
	Create(bus Bus, cfg C, opts ...O) (M, error)
}
