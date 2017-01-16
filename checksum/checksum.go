package checksum

import (
	"crypto/sha256"
	"errors"
	"io"
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

func Sum(rd io.Reader) (cks Hash, err error) {
	hash := sha256.New()
	_, err = io.Copy(hash, rd)
	if err != nil {
		return
	}
	hash.Sum(cks[:0])
	return
}

func SumBytes(b []byte) Hash {
	return sha256.Sum256(b)
}
