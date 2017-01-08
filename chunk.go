package scat

import "scat/checksum"

type Chunk interface {
	Num() int
	Data() []byte
	WithData([]byte) Chunk
	Hash() checksum.Hash
	SetHash(checksum.Hash)
	TargetSize() int
	SetTargetSize(int)
	Meta() Meta
}

type Meta interface {
	Get(interface{}) interface{}
	Set(_, _ interface{})
}

type chunk struct {
	num        int
	data       []byte
	hash       checksum.Hash
	targetSize int
	meta       meta
}

func NewChunk(num int, data []byte) Chunk {
	return &chunk{
		num:        num,
		data:       data,
		targetSize: len(data),
	}
}

func (c *chunk) Num() int {
	return c.num
}

func (c *chunk) Data() []byte {
	return c.data
}

func (c *chunk) WithData(d []byte) Chunk {
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

type meta map[interface{}]interface{}

func (m meta) Get(k interface{}) interface{} {
	return m[k]
}

func (m meta) Set(k, v interface{}) {
	m[k] = v
}
