package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/restic/chunker"

	"secdice/checksum"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func start() error {
	args := os.Args[1:]
	if len(args) != 1 {
		return errors.New("usage: split|join")
	}
	cmd := args[0]
	switch cmd {
	case "split":
		return split()
	case "join":
		return join()
	}
	return fmt.Errorf("unknown cmd: %s", cmd)
}

func split() (err error) {
	split := newSplitter(os.Stdin)
	ppool := newProcPool(2, procChain{
		newIndex(os.Stdout),
		&compress{},
		&localStore{"."},
	})
	err = process(split, ppool)
	if err != nil {
		return
	}
	return ppool.Finish()
}

func join() error {
	w := os.Stdout
	chain := []unprocessor{
		&localStore{"."},
		&compress{},
		&out{w},
	}
	scan := checksum.NewScanner(os.Stdin)
	num := 0
	for scan.Scan() {
		chunk := &Chunk{Num: num, Checksum: scan.Checksum}
		for _, p := range chain {
			err := p.Unprocess(chunk)
			if err != nil {
				return err
			}
		}
		num++
	}
	if err := scan.Err; err != nil {
		return err
	}
	return nil
}

type out struct {
	w io.Writer
}

func (out *out) Unprocess(c *Chunk) (err error) {
	_, err = out.w.Write(c.Data)
	return
}

func process(it chunkIterator, proc asyncProcessor) error {
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
			ch := proc.AsyncProcess(c)
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

type compress struct{}

func (*compress) Process(c *Chunk) (err error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(c.Data)))
	w := gzip.NewWriter(buf)
	_, err = w.Write(c.Data)
	if err != nil {
		return
	}
	err = w.Close()
	if err != nil {
		return
	}
	c.Data = buf.Bytes()
	return
}

func (*compress) Unprocess(c *Chunk) (err error) {
	r, err := gzip.NewReader(bytes.NewReader(c.Data))
	if err != nil {
		return
	}
	c.Data, err = ioutil.ReadAll(r)
	return
}

func (*compress) Finish() error {
	return nil
}

type index struct {
	w       io.Writer
	seen    map[checksum.Sum]struct{}
	seenMu  sync.Mutex
	order   []*checksum.Sum
	orderMu sync.Mutex
}

func newIndex(w io.Writer) *index {
	return &index{
		w:    w,
		seen: make(map[checksum.Sum]struct{}),
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

func (i *index) getSeen(cks checksum.Sum) (ok bool) {
	i.seenMu.Lock()
	defer i.seenMu.Unlock()
	_, ok = i.seen[cks]
	return
}

func (i *index) setSeen(cks checksum.Sum) {
	i.seenMu.Lock()
	defer i.seenMu.Unlock()
	i.seen[cks] = struct{}{}
}

func (i *index) setOrder(cks checksum.Sum, num int) {
	i.orderMu.Lock()
	defer i.orderMu.Unlock()
	if minLen := num + 1; len(i.order) < minLen {
		if cap(i.order) < minLen {
			resized := make([]*checksum.Sum, minLen, num*2+1)
			copy(resized, i.order)
			i.order = resized
		}
		i.order = i.order[:minLen]
	}
	i.order[num] = &cks
}

func (i *index) Finish() (err error) {
	for num, cks := range i.order {
		if cks == nil {
			return fmt.Errorf("missing chunk %d", num)
		}
	}
	for _, cks := range i.order {
		_, err = checksum.Write(i.w, *cks)
		if err != nil {
			return
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
	num     int // int for use as slice index
	chunk   *Chunk
	err     error
}

const (
	defaultMinSize = chunker.MinSize
	defaultMaxSize = chunker.MaxSize
	minMinSize     = 512 * 1024 // chunker.chunkerBufSize
)

func newSplitter(r io.Reader) *splitter {
	return newSplitterSize(r, defaultMinSize, defaultMaxSize)
}

func newSplitterSize(r io.Reader, minSize, maxSize uint) *splitter {
	if minSize < minMinSize {
		panic(fmt.Sprintf("min size must be >= %d bytes", minMinSize))
	}
	chunker := chunker.New(r, chunker.Pol(0x3DA3358B4DC173))
	chunker.MinSize = minSize
	chunker.MaxSize = maxSize
	return &splitter{
		chunker: chunker,
		buf:     make([]byte, maxSize),
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
	s.chunk = &Chunk{
		Num:  s.num,
		Data: data,
	}
	s.num++
	// Check for overflow: uint resets to 0, int resets to -minInt
	if s.num <= 0 {
		panic("overflow")
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
	Num      int
	Data     []byte
	Checksum checksum.Sum
}

type processor interface {
	Process(*Chunk) error
	Finish() error
}

type asyncProcessor interface {
	AsyncProcess(*Chunk) <-chan error
	Finish() error
}

type unprocessor interface {
	Unprocess(*Chunk) error
}

type localStore struct {
	Dir string
}

func (s *localStore) Process(c *Chunk) (err error) {
	path := s.path(c)
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

func (s *localStore) Unprocess(c *Chunk) (err error) {
	path := s.path(c)
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	c.Data, err = ioutil.ReadAll(f)
	return
}

func (s *localStore) path(c *Chunk) string {
	return filepath.Join(s.Dir, fmt.Sprintf("%x", c.Checksum))
}

type procChain []processor

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
	finish := func(p processor) {
		defer wg.Done()
		results <- p.Finish()
	}
	for _, p := range c {
		go finish(p)
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
	proc  processor
}

type procTask struct {
	chunk *Chunk
	res   chan<- error
}

func newProcPool(size int, proc processor) procPool {
	tasks := make(chan procTask)
	wg := &sync.WaitGroup{}
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func() {
			defer wg.Done()
			for task := range tasks {
				task.res <- proc.Process(task.chunk)
			}
		}()
	}
	return procPool{tasks: tasks, wg: wg, proc: proc}
}

func (p procPool) AsyncProcess(c *Chunk) <-chan error {
	res := make(chan error)
	p.tasks <- procTask{chunk: c, res: res}
	return res
}

func (p procPool) Finish() error {
	close(p.tasks)
	p.wg.Wait()
	return p.proc.Finish()
}
