package encryption

import (
	"math/rand"
	"time"
)

// using aes256, so 32 bytes-long key
const KEYLEN uint = 32
const CHARS string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Generates 32 pseudo-random bytes to use as a key
func Generate32AESkey() []byte {
	var generatedKey []byte

	rand.Seed(time.Now().UTC().UnixNano())
	// choosing "random" 32 bytes from CHARS
	for {
		if len(generatedKey) == int(KEYLEN) {
			break
		}
		randomIndex := rand.Intn(len(CHARS))
		generatedKey = append(generatedKey, CHARS[randomIndex])
	}

	return generatedKey
}
