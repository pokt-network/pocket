package testutil

// Concatenate appends the contents of multiple slices of any type (T) into a
// single slice of type T.
func Concatenate[T any](tt ...[]T) []T {
	var result []T
	for _, t := range tt {
		result = append(result, t...)
	}

	return result
}
