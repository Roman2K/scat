package secsplit

import "secsplit/checksum"

type Chunk struct {
	Num  int
	Data []byte
	Hash checksum.Hash
	Meta map[string]interface{}
}

func (c *Chunk) GetMeta(key string) interface{} {
	if c.Meta == nil {
		return nil
	}
	return c.Meta[key]
}

func (c *Chunk) SetMeta(key string, val interface{}) {
	if c.Meta == nil {
		c.Meta = make(map[string]interface{})
	}
	c.Meta[key] = val
}

type ChunkIterator interface {
	Next() bool
	Chunk() *Chunk
	Err() error
}
