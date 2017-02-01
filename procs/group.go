package procs

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/Roman2K/scat"
)

type group struct {
	size      int
	growing   map[int][]*scat.Chunk
	growingMu sync.Mutex
}

type Group interface {
	Proc
	ErrProc
}

func NewGroup(size int) Group {
	const min = 1
	if size < min {
		panic(fmt.Errorf("size must be >= %d", min))
	}
	return &group{
		size:    size,
		growing: make(map[int][]*scat.Chunk),
	}
}

func (g *group) ProcessErr(c *scat.Chunk, err error) <-chan Res {
	c.Meta().Set("err", err)
	return g.Process(c)
}

func (g *group) Process(c *scat.Chunk) <-chan Res {
	head, grouped, ok, err := g.build(c)
	ch := make(chan Res, 1)
	defer close(ch)
	if err != nil {
		ch <- Res{Chunk: c, Err: err}
	} else if ok {
		agg := scat.NewChunk(head, nil)
		agg.SetTargetSize(grouped[0].TargetSize())
		agg.Meta().Set("group", grouped)
		ch <- Res{Chunk: agg}
	}
	return ch
}

func (g *group) build(c *scat.Chunk) (
	head int, chunks []*scat.Chunk, ok bool, err error,
) {
	g.growingMu.Lock()
	defer g.growingMu.Unlock()
	head = c.Num() / g.size
	if _, ok := g.growing[head]; !ok {
		g.growing[head] = make([]*scat.Chunk, 0, g.size)
	}
	chunks = append(g.growing[head], c)
	have := len(chunks)
	if have < g.size {
		g.growing[head] = chunks
		return
	}
	delete(g.growing, head)
	if have != g.size {
		err = errors.New("accumulated too many chunks")
		return
	}
	num := func(i int) int {
		return chunks[i].Num()
	}
	sort.Slice(chunks, func(i, j int) bool {
		return num(i) < num(j)
	})
	if !contiguous(chunks) {
		err = errors.New("non-contiguous series")
	}
	ok = true
	return
}

func contiguous(chunks []*scat.Chunk) bool {
	for i := 1; i < len(chunks); i++ {
		if chunks[i].Num() != chunks[i-1].Num()+1 {
			return false
		}
	}
	return true
}

func (g *group) Finish() error {
	if len(g.growing) > 0 {
		return ErrShort
	}
	return nil
}
