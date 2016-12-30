package checksum

import (
	"crypto/sha256"
	"errors"
)

const size = sha256.Size

type Hash [size]byte

func (h *Hash) LoadSlice(s []byte) error {
	if len(s) != size {
		return errors.New("invalid hash length")
	}
	copy(h[:], s)
	return nil
}

func Sum(b []byte) Hash {
	return sha256.Sum256(b)
}
