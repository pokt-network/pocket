package utils

import "crypto/rand"

func GenerateRandomHash() []byte {
	buff := make([]byte, 8) // TODO(derrandz): revisit later. Decide on proper size for message hashes
	rand.Read(buff)
	return buff
}
