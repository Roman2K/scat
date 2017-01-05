package checksum

import (
	"crypto/sha256"
	"errors"
)

const Size = sha256.Size

type Hash [Size]byte

func (h *Hash) LoadSlice(s []byte) error {
	if len(s) != Size {
		return errors.New("invalid hash length")
	}
	copy(h[:], s)
	return nil
}

func Sum(b []byte) Hash {
	return sha256.Sum256(b)
}
