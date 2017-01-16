package procs

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

func checksumProcess(c scat.Chunk) (err error) {
	h, err := checksum.Sum(c.Data().Reader())
	if err != nil {
		return
	}
	c.SetHash(h)
	return
}

func checksumUnprocess(c scat.Chunk) (err error) {
	h, err := checksum.Sum(c.Data().Reader())
	if err != nil {
		return
	}
	if h != c.Hash() {
		return ErrIntegrityCheckFailed
	}
	return
}
