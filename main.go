package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/restic/chunker"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func start() error {
	spl := newSplitter(os.Stdin)
	return process(spl, []processor{
		&localStore{"."},
	})
}

func process(it chunkIterator, processors []processor) error {
	// TODO parallel
	for it.Next() {
		if err := it.Err(); err != nil {
			return err
		}
		chunk := it.Chunk()
		for _, p := range processors {
			if err := p.Process(chunk); err != nil {
				return err
			}
		}
	}
	return nil
}

type chunkIterator interface {
	Next() bool
	Chunk() *Chunk
	Err() error
}

type splitter struct {
	chunker *chunker.Chunker
	buf     []byte
	seen    map[checksum]struct{}
	chunk   *Chunk
	err     error
}

var _ chunkIterator = (*splitter)(nil)

func newSplitter(r io.Reader) *splitter {
	return &splitter{
		chunker: chunker.New(r, chunker.Pol(0x3DA3358B4DC173)),
		buf:     make([]byte, chunker.MaxSize),
		seen:    make(map[checksum]struct{}),
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
	sum := sha256.Sum256(c.Data)
	if _, ok := s.seen[sum]; ok {
		return s.Next()
	}
	s.seen[sum] = struct{}{}
	s.chunk = &Chunk{
		Data:     c.Data,
		Checksum: sum,
	}
	return true
}

func (s *splitter) Chunk() *Chunk {
	return s.chunk
}

func (s *splitter) Err() error {
	return s.err
}

type Chunk struct {
	Data     []byte
	Checksum checksum
}

type checksum [sha256.Size]byte

type processor interface {
	Process(*Chunk) error
}

type localStore struct {
	Dir string
}

var _ processor = (*localStore)(nil)

func (s *localStore) Process(c *Chunk) (err error) {
	path := filepath.Join(s.Dir, fmt.Sprintf("%x", c.Checksum))
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(c.Data)
	return
}
