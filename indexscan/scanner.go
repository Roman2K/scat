package indexscan

import (
	"io"

	ss "secsplit"
	"secsplit/checksum"
)

type scanner struct {
	scan  *checksum.Scanner
	num   int
	chunk *ss.Chunk
	err   error
}

func NewScanner(r io.Reader) ss.ChunkIterator {
	return &scanner{scan: checksum.NewScanner(r)}
}

func (s *scanner) Next() bool {
	ok := s.scan.Scan()
	if !ok {
		s.err = s.scan.Err
		return false
	}
	s.chunk = &ss.Chunk{
		Num:  s.num,
		Hash: s.scan.Hash,
	}
	s.num++ // TODO check overflow
	return true
}

func (s *scanner) Chunk() *ss.Chunk {
	return s.chunk
}

func (s *scanner) Err() error {
	return s.err
}
