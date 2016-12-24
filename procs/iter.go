package procs

import ss "secsplit"

type iter struct {
	ch     chan *ss.Chunk
	chunk  *ss.Chunk
	closed bool
}

func Iter(bufSize int) *iter {
	return &iter{ch: make(chan *ss.Chunk, bufSize)}
}

var _ ss.ChunkIterator = &iter{}

func (it *iter) Process(c *ss.Chunk) Res {
	return inplaceProcFunc(it.process).Process(c)
}

func (it *iter) process(c *ss.Chunk) error {
	it.ch <- c
	return nil
}

func (it *iter) Finish() error {
	if !it.closed {
		close(it.ch)
		it.closed = true
	}
	return nil
}

func (it *iter) Next() (ok bool) {
	it.chunk, ok = <-it.ch
	return
}

func (it *iter) Chunk() *ss.Chunk {
	return it.chunk
}

func (it *iter) Err() error {
	return nil
}
