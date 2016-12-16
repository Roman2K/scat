package checksum

import (
	"crypto/sha256"
	"fmt"
	"io"
)

type Hash [sha256.Size]byte

func Write(w io.Writer, s Hash) (int, error) {
	return fmt.Fprintf(w, "%x\n", s)
}

func Sum(b []byte) Hash {
	return sha256.Sum256(b)
}
