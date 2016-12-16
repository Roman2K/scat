package main

import (
	"bytes"
	"compress/gzip"
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
	index := newIndex(os.Stdout)
	ppool := newProcPool(8, procChain{
		index.Process,
		(&compress{}).Process,
		(&localStore{"out"}).Process,
	}.Process)
	err = process(split, ppool.Process)
	if err != nil {
		return
	}
	return parallel(index.Finish, ppool.Finish)
}

func parallel(fns ...func() error) (err error) {
	results := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(fns))
	call := func(fn func() error) {
		defer wg.Done()
		results <- fn()
	}
	for _, fn := range fns {
		go call(fn)
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

func join() error {
	w := os.Stdout
	iter := newIndexIterator(os.Stdin)
	// TODO proc pool, respect order from index iterator
	process := procChain{
		(&localStore{"out"}).Unprocess,
		(&compress{}).Unprocess,
		verify,
		(&out{w}).Unprocess,
	}.Process
	for iter.Next() {
		err := process(iter.Chunk())
		if err != nil {
			return err
		}
	}
	return iter.Err()
}

type indexIterator struct {
	r     io.Reader
	scan  *checksum.Scanner
	num   int
	chunk *Chunk
	err   error
}

func newIndexIterator(r io.Reader) *indexIterator {
	return &indexIterator{scan: checksum.NewScanner(r)}
}

func (it *indexIterator) Next() bool {
	ok := it.scan.Scan()
	if !ok {
		it.err = it.scan.Err
		return false
	}
	it.chunk = &Chunk{
		Num:  it.num,
		Hash: it.scan.Hash,
	}
	it.num++ // TODO check overflow
	return true
}

func (it *indexIterator) Chunk() *Chunk {
	return it.chunk
}

func (it *indexIterator) Err() error {
	return it.err
}

func verify(c *Chunk) error {
	if got := checksum.Sum(c.Data); got != c.Hash {
		return fmt.Errorf("integrity check failed for chunk #%d: got %x, want %x",
			c.Num, got, c.Hash)
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

func process(it chunkIterator, proc asyncProcFunc) error {
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
			ch := proc(c)
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

type compress struct {
	// TODO level
}

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

type index struct {
	w       io.Writer
	seen    map[checksum.Hash]struct{}
	seenMu  sync.Mutex
	order   []*checksum.Hash
	orderMu sync.Mutex
}

func newIndex(w io.Writer) *index {
	return &index{
		w:    w,
		seen: make(map[checksum.Hash]struct{}),
	}
}

func (i *index) Process(c *Chunk) error {
	c.Hash = checksum.Sum(c.Data)
	i.setOrder(c.Hash, c.Num)
	if i.getSeen(c.Hash) {
		c.Data = nil
		return nil
	}
	i.setSeen(c.Hash)
	return nil
}

func (i *index) getSeen(hash checksum.Hash) (ok bool) {
	i.seenMu.Lock()
	defer i.seenMu.Unlock()
	_, ok = i.seen[hash]
	return
}

func (i *index) setSeen(hash checksum.Hash) {
	i.seenMu.Lock()
	defer i.seenMu.Unlock()
	i.seen[hash] = struct{}{}
}

func (i *index) setOrder(hash checksum.Hash, num int) {
	i.orderMu.Lock()
	defer i.orderMu.Unlock()
	if minLen := num + 1; len(i.order) < minLen {
		if cap(i.order) < minLen {
			resized := make([]*checksum.Hash, minLen, num*2+1)
			copy(resized, i.order)
			i.order = resized
		}
		i.order = i.order[:minLen]
	}
	i.order[num] = &hash
}

func (i *index) Finish() (err error) {
	for num, hash := range i.order {
		if hash == nil {
			return fmt.Errorf("missing chunk %d", num)
		}
	}
	for _, hash := range i.order {
		_, err = checksum.Write(i.w, *hash)
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
	Num  int
	Data []byte
	Hash checksum.Hash
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
	return filepath.Join(s.Dir, fmt.Sprintf("%x", c.Hash))
}

type procFunc func(*Chunk) error
type asyncProcFunc func(*Chunk) <-chan error

type procChain []procFunc

func (c procChain) Process(chunk *Chunk) error {
	for _, process := range c {
		if err := process(chunk); err != nil {
			return err
		}
		if chunk.Data == nil {
			break
		}
	}
	return nil
}

type procPool struct {
	wg    *sync.WaitGroup
	tasks chan procTask
}

type procTask struct {
	chunk *Chunk
	res   chan<- error
}

func newProcPool(size int, proc procFunc) procPool {
	tasks := make(chan procTask)
	wg := &sync.WaitGroup{}
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func() {
			defer wg.Done()
			for task := range tasks {
				task.res <- proc(task.chunk)
			}
		}()
	}
	return procPool{wg: wg, tasks: tasks}
}

func (p procPool) Process(c *Chunk) <-chan error {
	res := make(chan error)
	p.tasks <- procTask{chunk: c, res: res}
	return res
}

func (p procPool) Finish() error {
	close(p.tasks)
	p.wg.Wait()
	return nil
}
