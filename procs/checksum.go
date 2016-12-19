package procs

import (
	"errors"
	ss "secsplit"
	"secsplit/checksum"
)

var errIntegrityCheckFailed = errors.New("checksum verification failed")

type Checksum struct{}

func (cks Checksum) Proc() Proc {
	return inplaceProcFunc(cks.process)
}

func (cks Checksum) Unproc() Proc {
	return inplaceProcFunc(cks.unprocess)
}

func (cks Checksum) process(c *ss.Chunk) error {
	c.Hash = checksum.Sum(c.Data)
	return nil
}

func (cks Checksum) unprocess(c *ss.Chunk) error {
	if checksum.Sum(c.Data) != c.Hash {
		return errIntegrityCheckFailed
	}
	return nil
}
