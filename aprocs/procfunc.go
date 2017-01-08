package aprocs

import "scat"

type ProcFunc func(scat.Chunk) <-chan Res

func (fn ProcFunc) Process(c scat.Chunk) <-chan Res {
	return fn(c)
}

func (ProcFunc) Finish() error {
	return nil
}

type InplaceFunc func(scat.Chunk) error

func (fn InplaceFunc) Process(c scat.Chunk) <-chan Res {
	ch := make(chan Res, 1)
	err := fn(c)
	ch <- Res{Chunk: c, Err: err}
	close(ch)
	return ch
}

func (InplaceFunc) Finish() error {
	return nil
}

type ChunkFunc func(scat.Chunk) (scat.Chunk, error)

func (fn ChunkFunc) Process(c scat.Chunk) <-chan Res {
	ch := make(chan Res, 1)
	new, err := fn(c)
	if err != nil {
		ch <- Res{Chunk: c, Err: err}
	} else {
		ch <- Res{Chunk: new}
	}
	close(ch)
	return ch
}

func (ChunkFunc) Finish() error {
	return nil
}

var (
	_ Proc = ProcFunc(nil)
	_ Proc = InplaceFunc(nil)
	_ Proc = ChunkFunc(nil)
)
