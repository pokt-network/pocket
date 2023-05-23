package generics_testutil

func GetKeys[K comparable, V any](keyMap map[K]V) []K {
	var (
		idx  = 0
		keys = make([]K, len(keyMap))
	)
	for key := range keyMap {
		keys[idx] = key
		idx++
	}
	return keys
}
