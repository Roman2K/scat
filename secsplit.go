package secsplit

import "secsplit/checksum"

type Chunk struct {
	Num  int
	Data []byte
	Hash checksum.Hash
}

type ChunkIterator interface {
	Next() bool
	Chunk() *Chunk
	Err() error
}
