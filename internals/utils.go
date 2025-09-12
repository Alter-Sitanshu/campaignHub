package internals

import (
	"crypto/rand"
	"log"
	"math/big"
)

func RandString(size int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var output string

	for range size {
		// pick a random index
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Printf("error generating random int: %v", err)
			return ""
		}
		output += string(letters[n.Int64()])
	}
	return output
}
