package meta

import "crypto/sha256"

type Split struct {
	Size   int64
	Sha256 [sha256.Size]byte
}
