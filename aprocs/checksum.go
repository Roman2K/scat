package aprocs

import (
	"errors"

	"scat"
	"scat/checksum"
)

var ErrIntegrityCheckFailed = errors.New("checksum verification failed")

var (
	ChecksumProc   Proc = InplaceFunc(checksumProcess)
	ChecksumUnproc Proc = InplaceFunc(checksumUnprocess)
)

func checksumProcess(c scat.Chunk) error {
	c.SetHash(checksum.Sum(c.Data()))
	return nil
}

func checksumUnprocess(c scat.Chunk) error {
	if checksum.Sum(c.Data()) != c.Hash() {
		return ErrIntegrityCheckFailed
	}
	return nil
}
