package indexscan

import (
	"fmt"
	"io"

	"secsplit/checksum"
)

func Write(w io.Writer, hash checksum.Hash, size int) (int, error) {
	return fmt.Fprintf(w, "%x %d\n", hash, size)
}
