package scat

import (
	"bytes"
	"io"
	"io/ioutil"

	"scat/checksum"
)

type Chunk interface {
	Num() int
	Data() Data
	WithData(Data) Chunk
	Hash() checksum.Hash
	SetHash(checksum.Hash)
	TargetSize() int
	SetTargetSize(int)
	Meta() Meta
}

type chunk struct {
	num        int
	data       Data
	hash       checksum.Hash
	targetSize int
	meta       meta
}

func NewChunk(num int, data Data) Chunk {
	if data == nil {
		data = BytesData(nil)
	}
	return &chunk{
		num:  num,
		data: data,
	}
}

func (c *chunk) Num() int {
	return c.num
}

func (c *chunk) Data() Data {
	return c.data
}

func (c *chunk) WithData(d Data) Chunk {
	dup := *c
	dup.data = d
	dup.meta = c.dupMeta()
	return &dup
}

func (c *chunk) Hash() checksum.Hash {
	return c.hash
}

func (c *chunk) SetHash(h checksum.Hash) {
	c.hash = h
}

func (c *chunk) TargetSize() int {
	return c.targetSize
}

func (c *chunk) SetTargetSize(s int) {
	c.targetSize = s
}

func (c *chunk) Meta() Meta {
	if c.meta == nil {
		c.meta = make(meta)
	}
	return c.meta
}

func (c *chunk) dupMeta() (dup meta) {
	dup = make(meta)
	for k, v := range c.meta {
		dup[k] = v
	}
	return
}

type Meta interface {
	Get(interface{}) interface{}
	Set(_, _ interface{})
}

type meta map[interface{}]interface{}

func (m meta) Get(k interface{}) interface{} {
	return m[k]
}

func (m meta) Set(k, v interface{}) {
	m[k] = v
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
