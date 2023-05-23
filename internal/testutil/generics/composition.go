package generics_testutil

// PipeTwoToOne threads two values of any type (T and U) through a pipeline of
// functions and returns the result of any type (U). Each function in the pipeline
// takes two arguments of type T and U, and returns a value of type U.
//
// Applies each function in the pipeline to the current value of U and the
// constant value of T, effectively "threading" the initial U value through the
// pipeline of functions.
//
// Does *not* mutate the original U value. Instead, it operates on a reference
// to U, ensuring that value types (non-pointer types) are not mutated.
//
// Returns the final value of U after it has been threaded through all the functions in the pipeline.
//
// Usage:
//
//	result := PipeTwo(initialT, initialU, func1, func2, func3)
//
// In this example, initialT and initialU are the initial values of T and U, and func1, func2, and func3
// are functions that take two arguments of type T and U and return a value of type U.
func PipeTwoToOne[T, U any](t T, u U, pipeline ...func(T, U) U) U {
	// NB: don't mutate potential value type `u` (i.e. non-pointer)
	uRef := u
	for _, fn := range pipeline {
		uRef = fn(t, uRef)
	}

	return u
}
