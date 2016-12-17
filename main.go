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
	index := &index{w: os.Stdout}

	ppool := newProcPool(8, procChain{
		inplace(computeChecksum).Process,
		inplace(newDedup().ProcessInplace).Process,
		splitChunk,
		inplace(computeChecksum).Process,
		inplace((&compress{}).ProcessInplace).Process,
		// (&paritySplit{data: 2, parity: 1}).Process,
		inplace((&localStore{"out"}).ProcessInplace).Process,
		index.Process,
	}.Process)

	err = process(split, ppool.Process)
	if err != nil {
		return
	}

	return parallel(index.Finish, ppool.Finish)
}

func splitChunk(c *Chunk) outChunk {
	const num = 2
	offset := len(c.Data) / num
	boundaries := [num][2]int{{0, offset}, {offset, len(c.Data)}}
	chunks := make([]*Chunk, len(boundaries))
	for i, bds := range boundaries {
		start, end := bds[0], bds[1]
		data := make([]byte, end-start)
		copy(data, c.Data[start:end])
		// TODO check overflow
		chunks[i] = &Chunk{Num: c.Num*num + i, Data: data}
	}
	return outChunk{out: chunks}
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
	iter := newIndexScanner(os.Stdin)
	// TODO proc pool, respect order from index iterator
	process := procChain{
		inplace((&localStore{"out"}).UnprocessInplace).Process,
		// inplace((&compress{}).UnprocessInplace).Process,
		inplace(verify).Process,
		inplace((&out{w}).UnprocessInplace).Process,
	}.Process
	for iter.Next() {
		res := process(iter.Chunk())
		if e := res.err; e != nil {
			return e
		}
	}
	return iter.Err()
}

type indexScanner struct {
	r     io.Reader
	scan  *checksum.Scanner
	num   int
	chunk *Chunk
	err   error
}

func newIndexScanner(r io.Reader) *indexScanner {
	return &indexScanner{scan: checksum.NewScanner(r)}
}

func (s *indexScanner) Next() bool {
	ok := s.scan.Scan()
	if !ok {
		s.err = s.scan.Err
		return false
	}
	s.chunk = &Chunk{
		Num:  s.num,
		Hash: s.scan.Hash,
	}
	s.num++ // TODO check overflow
	return true
}

func (s *indexScanner) Chunk() *Chunk {
	return s.chunk
}

func (s *indexScanner) Err() error {
	return s.err
}

func verify(c *Chunk) error {
	if checksum.Sum(c.Data) != c.Hash {
		return fmt.Errorf("integrity check failed for chunk %d")
	}
	return nil
}

// type paritySplit struct {
// 	rs reedsolomon.Encoder
// }

// func newParitySplit(data, parity int) *paritySplit {
// 	return &paritySplit{rs: reedsolomon.New(data, parity)}
// }

// func (ps *paritySplit) Process(c *Chunk) outChunk {
// 	shards, err := rs.Split(c.Data)
// 	if err != nil {
// 		return outChunk{err: err}
// 	}
// 	out := make([]*Chunk, len(shards))
// 	for i, shard := range shards {
// 		out[i] = shard
// 	}
// 	return outChunk{out: out}
// }

type out struct {
	w io.Writer
}

func (out *out) UnprocessInplace(c *Chunk) (err error) {
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
				res := <-ch
				results <- res.err
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

type inplace func(*Chunk) error

func (fn inplace) Process(c *Chunk) outChunk {
	err := fn(c)
	return outChunk{out: []*Chunk{c}, err: err}
}

type compress struct {
	// TODO level
}

func (*compress) ProcessInplace(c *Chunk) (err error) {
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

func (*compress) UnprocessInplace(c *Chunk) (err error) {
	r, err := gzip.NewReader(bytes.NewReader(c.Data))
	if err != nil {
		return
	}
	c.Data, err = ioutil.ReadAll(r)
	return
}

func computeChecksum(c *Chunk) error {
	c.Hash = checksum.Sum(c.Data)
	return nil
}

type dedup struct {
	seen   map[checksum.Hash]struct{}
	seenMu sync.Mutex
}

func newDedup() *dedup {
	return &dedup{
		seen: make(map[checksum.Hash]struct{}),
	}
}

func (d *dedup) ProcessInplace(c *Chunk) error {
	if d.getSeen(c.Hash) {
		c.Dup = true
	} else {
		d.setSeen(c.Hash)
	}
	return nil
}

func (d *dedup) getSeen(hash checksum.Hash) (ok bool) {
	d.seenMu.Lock()
	defer d.seenMu.Unlock()
	_, ok = d.seen[hash]
	return
}

func (d *dedup) setSeen(hash checksum.Hash) {
	d.seenMu.Lock()
	defer d.seenMu.Unlock()
	d.seen[hash] = struct{}{}
}

type index struct {
	w       io.Writer
	order   []*checksum.Hash
	orderMu sync.Mutex
}

func (i *index) Process(c *Chunk) outChunk {
	return outChunk{out: []*Chunk{c}}
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
	Dup  bool
}

type procFunc func(*Chunk) outChunk
type asyncProcFunc func(*Chunk) <-chan outChunk

type outChunk struct {
	out []*Chunk
	err error
}

type localStore struct {
	Dir string
}

func (s *localStore) ProcessInplace(c *Chunk) (err error) {
	path := s.path(c)
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(c.Data)
	return
}

func (s *localStore) UnprocessInplace(c *Chunk) (err error) {
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

type procChain []procFunc

func (chain procChain) Process(chunk *Chunk) outChunk {
	chunks := []*Chunk{chunk}
	for _, process := range chain {
		// TODO allocate len(chunks) * <max chunks output by this processor>
		out := make([]*Chunk, 0, len(chunks))
		for _, chunk := range chunks {
			res := process(chunk)
			if res.err != nil {
				return res
			}
			out = append(out, res.out...)
		}
		chunks = out
	}
	return outChunk{out: chunks}
}

type procPool struct {
	wg    *sync.WaitGroup
	tasks chan procTask
}

type procTask struct {
	chunk *Chunk
	res   chan<- outChunk
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

func (p procPool) Process(c *Chunk) <-chan outChunk {
	res := make(chan outChunk)
	p.tasks <- procTask{chunk: c, res: res}
	return res
}

func (p procPool) Finish() error {
	close(p.tasks)
	p.wg.Wait()
	return nil
}
