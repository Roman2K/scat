package procs

import "github.com/Roman2K/scat"

var (
	_ Proc = ProcFunc(nil)
	_ Proc = InplaceFunc(nil)
	_ Proc = ChunkFunc(nil)
	_ Proc = ChunkIterFunc(nil)
)

type ProcFunc func(*scat.Chunk) <-chan Res

func (fn ProcFunc) Process(c *scat.Chunk) <-chan Res {
	return fn(c)
}

func (ProcFunc) Finish() error {
	return nil
}

type InplaceFunc func(*scat.Chunk) error

func (fn InplaceFunc) Process(c *scat.Chunk) <-chan Res {
	ch := make(chan Res, 1)
	defer close(ch)
	err := fn(c)
	ch <- Res{Chunk: c, Err: err}
	return ch
}

func (InplaceFunc) Finish() error {
	return nil
}

type ChunkFunc func(*scat.Chunk) (*scat.Chunk, error)

func (fn ChunkFunc) Process(c *scat.Chunk) <-chan Res {
	ch := make(chan Res, 1)
	defer close(ch)
	if new, err := fn(c); err != nil {
		ch <- Res{Chunk: c, Err: err}
	} else {
		ch <- Res{Chunk: new}
	}
	return ch
}

func (ChunkFunc) Finish() error {
	return nil
}

type ChunkIterFunc func(*scat.Chunk) scat.ChunkIter

func (fn ChunkIterFunc) Process(c *scat.Chunk) <-chan Res {
	iter := fn(c)
	ch := make(chan Res)
	go func() {
		defer close(ch)
		for iter.Next() {
			ch <- Res{Chunk: iter.Chunk()}
		}
		if err := iter.Err(); err != nil {
			ch <- Res{Chunk: c, Err: err}
		}
	}()
	return ch
}

func (ChunkIterFunc) Finish() error {
	return nil
}
