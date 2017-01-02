package index

import (
	"fmt"
	"io"
	"secsplit/checksum"

	ss "secsplit"
)

type scanner struct {
	r       io.Reader
	hashBuf []byte
	num     int
	chunk   *ss.Chunk
	err     error
}

func NewScanner(r io.Reader) ss.ChunkIterator {
	return &scanner{
		r:       r,
		hashBuf: make([]byte, len(checksum.Hash{})),
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
	chunk := &ss.Chunk{
		Num:  s.num,
		Size: size,
	}
	err = chunk.Hash.LoadSlice(s.hashBuf)
	if err != nil {
		return
	}
	s.chunk = chunk
	s.num++
	return
}

func (s *scanner) Chunk() *ss.Chunk {
	return s.chunk
}

func (s *scanner) Err() error {
	return s.err
}
