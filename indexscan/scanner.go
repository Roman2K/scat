package indexscan

import (
	"errors"
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

func (s *scanner) scan() error {
	var size int
	n, err := fmt.Fscanf(s.r, "%x %d\n", &s.hashBuf, &size)
	if err != nil {
		return err
	}
	if n != 2 {
		return fmt.Errorf("failed to read index line")
	}
	chunk := &ss.Chunk{
		Num:  s.num,
		Size: size,
	}
	n = copy(chunk.Hash[:], s.hashBuf)
	if n != len(chunk.Hash) {
		return errors.New("invalid hash length")
	}
	s.chunk = chunk
	s.num++
	return nil
}

func (s *scanner) Chunk() *ss.Chunk {
	return s.chunk
}

func (s *scanner) Err() error {
	return s.err
}
