package procs

import (
	ss "secsplit"
	"secsplit/checksum"
)

type MemStore map[checksum.Hash][]byte

var _ ProcUnprocer = MemStore{}

func (ms MemStore) Proc() Proc {
	return inplaceProcFunc(ms.process)
}

func (ms MemStore) Unproc() Proc {
	return inplaceProcFunc(ms.unprocess)
}

func (ms MemStore) process(c *ss.Chunk) error {
	ms[c.Hash] = copyBytes(c.Data)
	return nil
}

func (ms MemStore) unprocess(c *ss.Chunk) error {
	c.Data = copyBytes(ms[c.Hash])
	return nil
}

func copyBytes(b []byte) (bcopy []byte) {
	bcopy = make([]byte, len(b))
	copy(bcopy, b)
	return
}
