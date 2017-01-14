package split

import (
	"fmt"
	"io"

	"github.com/restic/chunker"

	"scat"
)

const (
	defaultMinSize = chunker.MinSize
	defaultMaxSize = chunker.MaxSize
	minMinSize     = 512 * 1024 // chunker.chunkerBufSize
	pol            = chunker.Pol(0x3DA3358B4DC173)
)

type splitter struct {
	chunker *chunker.Chunker
	buf     []byte
	num     int // int for use as slice index
	chunk   scat.Chunk
	err     error
}

func NewSplitter(r io.Reader) scat.ChunkIter {
	return NewSplitterSize(r, defaultMinSize, defaultMaxSize)
}

func NewSplitterSize(r io.Reader, min, max uint) scat.ChunkIter {
	if min < minMinSize {
		panic(fmt.Sprintf("min size must be >= %d bytes", minMinSize))
	}
	chunker := chunker.NewWithBoundaries(r, pol, min, max)
	return &splitter{
		chunker: chunker,
		buf:     make([]byte, max),
	}
}

func (s *splitter) Next() bool {
	c, err := s.chunker.Next(s.buf)
	if err != nil {
		if err == io.EOF {
			err = nil
		}
		s.err = err
		return false
	}
	data := make([]byte, len(c.Data))
	copy(data, c.Data)
	s.chunk = scat.NewChunk(s.num, data)
	s.num++
	// Check for overflow: uint resets to 0, int resets to -minInt
	if s.num <= 0 {
		panic("overflow")
	}
	return true
}

func (s *splitter) Chunk() scat.Chunk {
	return s.chunk
}

func (s *splitter) Err() error {
	return s.err
}
