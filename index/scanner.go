package index

import (
	"fmt"
	"io"

	"scat"
	"scat/checksum"
)

type scanner struct {
	r       io.Reader
	hashBuf []byte
	num     int
	chunk   *scat.Chunk
	err     error
}

func NewScanner(num int, r io.Reader) scat.ChunkIter {
	return &scanner{
		r:       r,
		hashBuf: make([]byte, len(checksum.Hash{})),
		num:     num,
	}
}

func (s *scanner) Next() bool {
	err := s.scan()
	if err != nil {
		if err == io.EOF {
			err = nil
		}
		s.err = err
		return false
	}
	return true
}

func (s *scanner) scan() (err error) {
	var size int
	n, err := fmt.Fscanf(s.r, "%x %d\n", &s.hashBuf, &size)
	if err != nil {
		return
	}
	if n != 2 {
		return fmt.Errorf("invalid index line")
	}
	chunk := scat.NewChunk(s.num, nil)
	chunk.SetTargetSize(size)
	hash := checksum.Hash{}
	err = hash.LoadSlice(s.hashBuf)
	if err != nil {
		return
	}
	chunk.SetHash(hash)
	s.chunk = chunk
	s.num++
	return
}

func (s *scanner) Chunk() *scat.Chunk {
	return s.chunk
}

func (s *scanner) Err() error {
	return s.err
}
