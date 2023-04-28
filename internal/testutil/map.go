package testutil

func GetKeys[K comparable, V any](keyMap map[K]V) (keys []K) {
	for key := range keyMap {
		keys = append(keys, key)
	}
	return keys
}
