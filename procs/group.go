package procs

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	ss "secsplit"
)

type group struct {
	size      int
	growing   map[int][]*ss.Chunk
	growingMu sync.Mutex
}

func Group(size int) *group {
	const min = 1
	if size < min {
		panic(fmt.Sprintf("size must be >= %d", min))
	}
	return &group{
		size:    size,
		growing: make(map[int][]*ss.Chunk),
	}
}

func (g *group) ProcessErr(c *ss.Chunk, err error) Res {
	c.SetMeta("err", err)
	return g.Process(c)
}

func (g *group) Process(c *ss.Chunk) Res {
	head, grouped, ok, err := g.build(c)
	chunks := make([]*ss.Chunk, 0, 1)
	if ok {
		agg := *grouped[0]
		agg.Num = head
		agg.SetMeta("group", grouped)
		chunks = append(chunks, &agg)
	}
	return Res{Chunks: chunks, Err: err}
}

func (g *group) build(c *ss.Chunk) (
	head int, chunks []*ss.Chunk, ok bool, err error,
) {
	g.growingMu.Lock()
	defer g.growingMu.Unlock()
	head = c.Num / g.size
	if _, ok := g.growing[head]; !ok {
		g.growing[head] = make([]*ss.Chunk, 0, g.size)
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
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Num < chunks[j].Num
	})
	if !contiguous(chunks) {
		err = errors.New("non-contiguous series")
	}
	ok = true
	return
}

func contiguous(chunks []*ss.Chunk) bool {
	for i := 1; i < len(chunks); i++ {
		if chunks[i].Num != chunks[i-1].Num+1 {
			return false
		}
	}
	return true
}

func (g *group) Finish() (err error) {
	if len(g.growing) > 0 {
		err = ErrShort
	}
	return
}
