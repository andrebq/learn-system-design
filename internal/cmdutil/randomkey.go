package cmdutil

import (
	"crypto/rand"
	"io"
)

// RandomKey generates a new random sequence of bytes using
// crypto/rand
func RandomKey(out []byte) {
	_, err := io.ReadFull(rand.Reader, out)
	if err != nil {
		panic(err)
	}
}
