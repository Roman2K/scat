package scat

type ChunkIter interface {
	Next() bool
	Chunk() Chunk
	Err() error
}
