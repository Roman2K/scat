package stores

import (
	"errors"
	"scat"
	"scat/checksum"
	"scat/procs"
	"sync"
)

type Mem struct {
	data   memMap
	dataMu sync.RWMutex
}

type memMap map[checksum.Hash][]byte

var _ Store = (*Mem)(nil)

func NewMem() *Mem {
	return &Mem{
		data: make(memMap),
	}
}

func (s *Mem) Proc() procs.Proc {
	return procs.InplaceFunc(s.process)
}

func (s *Mem) process(c *scat.Chunk) (err error) {
	b, err := c.Data().Bytes()
	if err != nil {
		return
	}
	s.SetData(c.Hash(), b)
	return nil
}

func (s *Mem) Unproc() procs.Proc {
	return procs.ChunkFunc(s.unprocess)
}

func (s *Mem) unprocess(c *scat.Chunk) (*scat.Chunk, error) {
	s.dataMu.RLock()
	data, ok := s.data[c.Hash()]
	s.dataMu.RUnlock()
	if !ok {
		return nil, procs.MissingDataError{errors.New("no stored data")}
	}
	dup := make(scat.BytesData, len(data))
	copy(dup, data)
	return c.WithData(dup), nil
}

func (s *Mem) SetData(hash checksum.Hash, b []byte) {
	s.dataMu.Lock()
	defer s.dataMu.Unlock()
	s.data[hash] = b
}

func (s *Mem) Ls() ([]LsEntry, error) {
	s.dataMu.RLock()
	defer s.dataMu.RUnlock()
	entries := make([]LsEntry, 0, len(s.data))
	for hash, data := range s.data {
		e := LsEntry{Hash: hash, Size: int64(len(data))}
		entries = append(entries, e)
	}
	return entries, nil
}
