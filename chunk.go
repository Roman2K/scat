package scat

import (
	"bytes"
	"io"
	"io/ioutil"
	"sync"

	"scat/checksum"
)

type Chunk struct {
	num        int
	data       Data
	hash       checksum.Hash
	targetSize int
	meta       *meta
}

func NewChunk(num int, data Data) *Chunk {
	if data == nil {
		data = BytesData(nil)
	}
	return &Chunk{
		num:  num,
		data: data,
	}
}

func (c *Chunk) Num() int {
	return c.num
}

func (c *Chunk) Data() Data {
	return c.data
}

func (c *Chunk) WithData(d Data) *Chunk {
	dup := *c
	dup.data = d
	if dup.meta != nil {
		dup.meta = c.meta.dup()
	}
	return &dup
}

func (c *Chunk) Hash() checksum.Hash {
	return c.hash
}

func (c *Chunk) SetHash(h checksum.Hash) {
	c.hash = h
}

func (c *Chunk) TargetSize() int {
	return c.targetSize
}

func (c *Chunk) SetTargetSize(s int) {
	c.targetSize = s
}

func (c *Chunk) Meta() Meta {
	if c.meta == nil {
		c.meta = newMeta()
	}
	return c.meta
}

type Meta interface {
	Get(interface{}) interface{}
	Set(_, _ interface{})
}

type meta struct {
	m  metaMap
	mu sync.RWMutex
}

type metaMap map[interface{}]interface{}

func newMeta() *meta {
	return &meta{m: make(metaMap)}
}

func (m *meta) Get(k interface{}) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.m[k]
}

func (m *meta) Set(k, v interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[k] = v
}

func (m *meta) dup() (dup *meta) {
	dup = newMeta()
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.m {
		dup.m[k] = v
	}
	return
}

type Data interface {
	Reader() io.Reader
	Bytes() ([]byte, error)
}

type Sizer interface {
	Size() int
}

type sizedData interface {
	Data
	Sizer
}

type BytesData []byte

var _ sizedData = BytesData{}

func (b BytesData) Reader() io.Reader {
	return bytes.NewReader([]byte(b))
}

func (b BytesData) Bytes() ([]byte, error) {
	return []byte(b), nil
}

func (b BytesData) Size() int {
	return len(b)
}

type readerData struct {
	r        io.Reader
	onceChan chan struct{}
}

func NewReaderData(r io.Reader) Data {
	return readerData{
		r:        r,
		onceChan: make(chan struct{}, 1),
	}
}

func (r readerData) Reader() (reader io.Reader) {
	r.once(func() {
		reader = r.r
	})
	return
}

func (r readerData) Bytes() (b []byte, err error) {
	r.once(func() {
		b, err = ioutil.ReadAll(r.r)
	})
	return
}

func (r readerData) once(fn func()) {
	select {
	case r.onceChan <- struct{}{}:
		fn()
	default:
		panic("reader data can only be read once")
	}
}
