package testutil

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
)

func RandomSHA1() string {
	buf := make([]byte, 16)

	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}

	hash := sha1.New()

	if _, err := hash.Write(buf); err != nil {
		panic(err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}
