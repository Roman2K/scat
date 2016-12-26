package aprocs_test

import (
	"sync"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
)

func TestChain(t *testing.T) {
	a := aprocs.InplaceProcFunc(func(c *ss.Chunk) error {
		c.Data = append(c.Data, 'a')
		return nil
	})
	b := aprocs.InplaceProcFunc(func(c *ss.Chunk) error {
		c.Data = append(c.Data, 'b')
		return nil
	})
	chain := aprocs.NewChain([]aprocs.Proc{a, b})
	ch := chain.Process(&ss.Chunk{Data: []byte{'x'}})
	res := <-ch
	_, ok := <-ch
	assert.False(t, ok)
	assert.NoError(t, res.Err)
	assert.Equal(t, "xab", string(res.Chunk.Data))
}

func TestChainEndProc(t *testing.T) {
	finals := make(map[*ss.Chunk][]int)
	ends := make(map[*ss.Chunk][]int)
	mu := sync.Mutex{}
	a := enderProc{
		proc: aprocs.ProcFunc(func(*ss.Chunk) <-chan aprocs.Res {
			ch := make(chan aprocs.Res, 1)
			ch <- aprocs.Res{Chunk: &ss.Chunk{Num: 11}}
			close(ch)
			return ch
		}),
		onFinal: func(c, final *ss.Chunk) error {
			mu.Lock()
			defer mu.Unlock()
			finals[c] = append(finals[c], final.Num)
			return nil
		},
		onEnd: func(c *ss.Chunk) error {
			mu.Lock()
			defer mu.Unlock()
			ends[c] = append(ends[c], finals[c]...)
			return nil
		},
	}
	b := aprocs.ProcFunc(func(*ss.Chunk) <-chan aprocs.Res {
		ch := make(chan aprocs.Res, 2)
		ch <- aprocs.Res{Chunk: &ss.Chunk{Num: 22}}
		ch <- aprocs.Res{Chunk: &ss.Chunk{Num: 33}}
		close(ch)
		return ch
	})
	chain := aprocs.NewChain([]aprocs.Proc{a, b})
	chunk := &ss.Chunk{Num: 0}
	ch := chain.Process(chunk)
	for range ch {
	}
	assert.Equal(t, 1, len(finals))
	assert.Equal(t, []int{22, 33}, finals[chunk])
	assert.Equal(t, []int{22, 33}, ends[chunk])
}

type enderProc struct {
	proc    aprocs.Proc
	onFinal func(*ss.Chunk, *ss.Chunk) error
	onEnd   func(*ss.Chunk) error
}

type ender interface {
	aprocs.Proc
	aprocs.EndProc
}

var _ ender = enderProc{}

func (e enderProc) Process(c *ss.Chunk) <-chan aprocs.Res {
	return e.proc.Process(c)
}

func (e enderProc) Finish() error {
	return e.proc.Finish()
}

func (e enderProc) ProcessFinal(c, final *ss.Chunk) error {
	return e.onFinal(c, final)
}

func (e enderProc) ProcessEnd(c *ss.Chunk) error {
	return e.onEnd(c)
}
