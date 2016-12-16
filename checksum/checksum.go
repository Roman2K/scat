package checksum

import (
	"crypto/sha256"
	"fmt"
	"io"
)

type Sum [sha256.Size]byte

func Write(w io.Writer, s Sum) (int, error) {
	return fmt.Fprintf(w, "%x\n", s)
}
