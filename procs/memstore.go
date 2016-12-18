package procs

import (
	ss "secsplit"
	"secsplit/checksum"
)

type memStore struct {
	data map[checksum.Hash][]byte
}

func MemStore() ProcUnprocer {
	return &memStore{
		data: make(map[checksum.Hash][]byte),
	}
}

func (ms *memStore) Proc() Proc {
	return inplaceProcFunc(ms.process)
}

func (ms *memStore) Unproc() Proc {
	return inplaceProcFunc(ms.unprocess)
}

func (ms *memStore) process(c *ss.Chunk) error {
	ms.data[c.Hash] = copyBytes(c.Data)
	return nil
}

func (ms *memStore) unprocess(c *ss.Chunk) error {
	c.Data = copyBytes(ms.data[c.Hash])
	return nil
}

func copyBytes(b []byte) (bcopy []byte) {
	bcopy = make([]byte, len(b))
	copy(bcopy, b)
	return
}
