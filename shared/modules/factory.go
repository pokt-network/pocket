package modules

// ModuleFactoryWithOptions implements a `#Create()` factory method which takes
// a variadic `ModuleOption` argument(s) and returns a `Module`and an error.
type ModuleFactoryWithOptions FactoryWithOptions[Module, ModuleOption]

// FactoryWithRequired implements a `#Create()` factory method which takes a
// required "config" argument of type K and returns a value of type T and an error.
type FactoryWithRequired[T interface{}, K interface{}] interface {
	Create(bus Bus, cfg K) (T, error)
}

// FactoryWithOptions implements a `#Create()` factory method which takes a
// variadic "optional" argument(s) of type O and returns a value of type T
// and an error.
type FactoryWithOptions[T interface{}, O interface{}] interface {
	Create(bus Bus, opts ...O) (T, error)
}

// FactoryWithRequiredAndOptions implements a `#Create()` factory method which
// takes both a required "config" argument of type K and a variadic "optional"
// argument(s) of type O and returns a value of type T and an error.
type FactoryWithRequiredAndOptions[T interface{}, K interface{}, O interface{}] interface {
	Create(bus Bus, cfg K, opts ...O) (T, error)
}
