package checksum

import "crypto/sha256"

type Hash [sha256.Size]byte

func Sum(b []byte) Hash {
	return sha256.Sum256(b)
}
