package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/restic/chunker"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func start() error {
	f, err := os.Open("/Users/roman/tmp/100m")
	if err != nil {
		return err
	}
	defer f.Close()

	splitters := newMultiChunkIt([]chunkIterator{
		newSplitter(os.Stdin),
		// newSplitter(f),
	})
	// splitters := newSplitter(f)
	defer splitters.Close()

	procPool := newProcPool([]processor{
		newIndex(),
		&localStore{"."},
		// &stats{start: time.Now()},
	})
	defer procPool.Close()

	return process(splitters, procPool)
}

type multiChunkIt struct {
	chunks chan *Chunk
	errors chan error
	done   chan struct{}
	chunk  *Chunk
	err    error
}

var _ chunkIterator = (*multiChunkIt)(nil)

func newMultiChunkIt(its []chunkIterator) *multiChunkIt {
	chunks := make(chan *Chunk)
	errors := make(chan error)
	done := make(chan struct{})

	wg := &sync.WaitGroup{}
	wg.Add(len(its))

	sendChunks := func(it chunkIterator) {
		defer wg.Done()
		for it.Next() {
			select {
			case chunks <- it.Chunk():
			case <-done:
				return
			}
		}
		if err := it.Err(); err != nil {
			errors <- err
		}
	}

	for _, it := range its {
		go sendChunks(it)
	}

	go func() {
		defer close(chunks)
		defer close(errors)
		wg.Wait()
	}()

	return &multiChunkIt{
		chunks: chunks,
		errors: errors,
		done:   done,
	}
}

func (mit *multiChunkIt) Next() bool {
	chunksDone, errsDone := false, false
	for !chunksDone || !errsDone {
		select {
		case chunk, ok := <-mit.chunks:
			if !ok {
				chunksDone = true
				continue
			}
			mit.chunk = chunk
			return true
		case err, ok := <-mit.errors:
			if !ok {
				errsDone = true
				continue
			}
			mit.err = err
			close(mit.done)
			return false
		}
	}
	return false
}

func (mit *multiChunkIt) Chunk() *Chunk {
	return mit.chunk
}

func (mit *multiChunkIt) Err() error {
	return mit.err
}

func (mit *multiChunkIt) Close() {
	select {
	case <-mit.done:
	default:
		close(mit.done)
	}
	for range mit.chunks {
	}
	for range mit.errors {
	}
}

// type stats struct {
// 	start time.Time
// 	n     float64
// }

// func (s *stats) Process(c *Chunk) error {
// 	if s.n > 0 {
// 		fmt.Printf("\r")
// 	}
// 	s.n += float64(len(c.Data)) / 1024.0 / 1024.0
// 	speed := s.n / time.Now().Sub(s.start).Seconds()
// 	fmt.Printf("speed: %d MB/s", uint64(speed))
// 	return nil
// }

func process(it chunkIterator, ppool procPool) error {
	chunks := make(chan *Chunk)
	results := make(chan error)
	done := make(chan struct{})
	resultSends := sync.WaitGroup{}

	resultSends.Add(1)
	go func() {
		defer resultSends.Done()
		defer close(chunks)
		for it.Next() {
			select {
			case chunks <- it.Chunk():
			case <-done:
				return
			}
		}
		results <- it.Err()
	}()

	resultSends.Add(1)
	go func() {
		defer resultSends.Done()
		for c := range chunks {
			ch := ppool.Process(c)
			resultSends.Add(1)
			go func() {
				defer resultSends.Done()
				err := <-ch
				results <- err
			}()
		}
	}()

	go func() {
		defer close(results)
		resultSends.Wait()
	}()

	collect := func() error {
		defer close(done)
		for err := range results {
			if err != nil {
				return err
			}
		}
		return nil
	}

	err := collect()
	for range results {
	}
	return err
}

type index struct {
	seen   map[checksum]struct{}
	seenMu sync.Mutex
	// TODO order []*checksum
}

func newIndex() *index {
	return &index{
		seen: make(map[checksum]struct{}),
	}
}

func (i *index) Process(c *Chunk) error {
	c.Checksum = sha256.Sum256(c.Data)
	if i.getSeen(c.Checksum) {
		c.Data = nil
		return nil
	}
	// TODO i.order = append(i.order, &cks)
	return nil
}

func (i *index) getSeen(cks checksum) (ok bool) {
	i.seenMu.Lock()
	defer i.seenMu.Unlock()
	_, ok = i.seen[cks]
	return
}

func (i *index) setSeen(cks checksum) {
	i.seenMu.Lock()
	defer i.seenMu.Unlock()
	i.seen[cks] = struct{}{}
}

type chunkIterator interface {
	Next() bool
	Chunk() *Chunk
	Err() error
	Close()
}

type splitter struct {
	chunker *chunker.Chunker
	buf     []byte
	num     uint
	chunk   *Chunk
	err     error
}

var _ chunkIterator = (*splitter)(nil)

func newSplitter(r io.Reader) *splitter {
	return &splitter{
		chunker: chunker.New(r, chunker.Pol(0x3DA3358B4DC173)),
		buf:     make([]byte, chunker.MaxSize),
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
	s.chunk = &Chunk{
		Num:  s.num,
		Data: c.Data,
	}
	s.num++ // TODO check overflow
	return true
}

func (s *splitter) Chunk() *Chunk {
	return s.chunk
}

func (s *splitter) Err() error {
	return s.err
}

func (s *splitter) Close() {
}

type Chunk struct {
	Num      uint
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
	return nil
	path := filepath.Join(s.Dir, fmt.Sprintf("%x", c.Checksum))
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(c.Data)
	return
}

type procPool struct {
	wg    *sync.WaitGroup
	tasks chan procTask
}

type procTask struct {
	chunk *Chunk
	res   chan error
}

func newProcPool(procs []processor) procPool {
	const nworkers = 1

	tasks := make(chan procTask)
	wg := &sync.WaitGroup{}
	wg.Add(nworkers)

	for i := 0; i < nworkers; i++ {
		go func() {
			defer wg.Done()
			for task := range tasks {
				for _, p := range procs {
					if err := p.Process(task.chunk); err != nil {
						task.res <- err
						break
					}
				}
				task.res <- nil
			}
		}()
	}

	return procPool{tasks: tasks, wg: wg}
}

func (p procPool) Process(c *Chunk) <-chan error {
	res := make(chan error)
	p.tasks <- procTask{chunk: c, res: res}
	return res
}

func (p procPool) Close() {
	close(p.tasks)
	p.wg.Wait()
}
