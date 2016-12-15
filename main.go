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

func start() (err error) {
	split := newSplitter(os.Stdin)

	chain := procChain{
		newIndex(),
		&localStore{"."},
	}

	procPool := newProcPool(chain)
	defer procPool.Close()

	err = process(split, procPool)
	if err != nil {
		return
	}

	return chain.Finish()
}

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
	seen    map[checksum]struct{}
	seenMu  sync.Mutex
	order   []*checksum
	orderMu sync.Mutex
}

func newIndex() *index {
	return &index{
		seen: make(map[checksum]struct{}),
	}
}

func (i *index) Process(c *Chunk) error {
	c.Checksum = sha256.Sum256(c.Data)
	i.setOrder(c.Checksum, c.Num)
	if i.getSeen(c.Checksum) {
		c.Data = nil
		return nil
	}
	i.setSeen(c.Checksum)
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

func (i *index) setOrder(cks checksum, num int) {
	i.orderMu.Lock()
	defer i.orderMu.Unlock()
	if minLen := num + 1; len(i.order) < minLen {
		if cap(i.order) < minLen {
			resized := make([]*checksum, minLen, num*2+1)
			copy(resized, i.order)
			i.order = resized
		}
		i.order = i.order[:minLen]
	}
	i.order[num] = &cks
}

func (i *index) Finish() error {
	var w io.Writer = os.Stdout
	for num, cks := range i.order {
		if cks == nil {
			return fmt.Errorf("missing chunk %d", num)
		}
	}
	for _, cks := range i.order {
		fmt.Fprintf(w, "%x\n", *cks)
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
	num     int // int for use as slice index
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

type Chunk struct {
	Num      int
	Data     []byte
	Checksum checksum
}

type checksum [sha256.Size]byte

type processor interface {
	Process(*Chunk) error
	Finish() error
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

func (s *localStore) Finish() error {
	return nil
}

type procChain []processor

var _ processor = procChain(nil)

func (c procChain) Process(chunk *Chunk) error {
	for _, p := range c {
		if err := p.Process(chunk); err != nil {
			return err
		}
		if chunk.Data == nil {
			break
		}
	}
	return nil
}

func (c procChain) Finish() (err error) {
	results := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(c))
	closeProc := func(p processor) {
		defer wg.Done()
		results <- p.Finish()
	}
	for _, p := range c {
		go closeProc(p)
	}
	go func() {
		defer close(results)
		wg.Wait()
	}()
	for e := range results {
		if e != nil && err == nil {
			err = e
		}
	}
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

func newProcPool(proc processor) procPool {
	const nworkers = 1
	tasks := make(chan procTask)
	wg := &sync.WaitGroup{}
	wg.Add(nworkers)
	for i := 0; i < nworkers; i++ {
		go func() {
			defer wg.Done()
			for task := range tasks {
				task.res <- proc.Process(task.chunk)
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
