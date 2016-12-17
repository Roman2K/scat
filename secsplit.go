package secsplit

import "secsplit/checksum"

type Chunk struct {
	Num  int
	Data []byte
	Hash checksum.Hash
	Dup  bool
}

type ChunkIterator interface {
	Next() bool
	Chunk() *Chunk
	Err() error
}
