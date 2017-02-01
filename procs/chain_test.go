package procs_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/procs"
	"github.com/Roman2K/scat/testutil"
	assert "github.com/stretchr/testify/require"
)

func TestChain(t *testing.T) {
	a := procs.ChunkFunc(func(c *scat.Chunk) (*scat.Chunk, error) {
		return appendData(c, 'a')
	})
	b := procs.ChunkFunc(func(c *scat.Chunk) (*scat.Chunk, error) {
		return appendData(c, 'b')
	})
	chain := procs.Chain{a, b}
	ch := chain.Process(scat.NewChunk(0, scat.BytesData{'x'}))
	res := <-ch
	_, ok := <-ch
	assert.False(t, ok)
	assert.NoError(t, res.Err)
	bytes, err := res.Chunk.Data().Bytes()
	assert.NoError(t, err)
	assert.Equal(t, "xab", string(bytes))
}

func appendData(c *scat.Chunk, b byte) (*scat.Chunk, error) {
	bytes, err := c.Data().Bytes()
	if err != nil {
		return nil, err
	}
	return c.WithData(append(scat.BytesData(bytes), b)), nil
}

func TestChainEndProc(t *testing.T) {
	finals := make(map[*scat.Chunk][]int)
	ends := make(map[*scat.Chunk][]int)
	mu := sync.Mutex{}
	a := enderProc{
		proc: procs.ProcFunc(func(*scat.Chunk) <-chan procs.Res {
			ch := make(chan procs.Res, 1)
			defer close(ch)
			ch <- procs.Res{Chunk: scat.NewChunk(11, nil)}
			return ch
		}),
		onFinal: func(c, final *scat.Chunk) error {
			mu.Lock()
			defer mu.Unlock()
			finals[c] = append(finals[c], final.Num())
			return nil
		},
		onEnd: func(c *scat.Chunk) error {
			mu.Lock()
			defer mu.Unlock()
			ends[c] = append(ends[c], finals[c]...)
			return nil
		},
	}
	b := procs.ProcFunc(func(*scat.Chunk) <-chan procs.Res {
		ch := make(chan procs.Res, 2)
		defer close(ch)
		ch <- procs.Res{Chunk: scat.NewChunk(22, nil)}
		ch <- procs.Res{Chunk: scat.NewChunk(33, nil)}
		return ch
	})
	chain := procs.Chain{a, b}
	chunk := scat.NewChunk(0, nil)
	ch := chain.Process(chunk)
	for range ch {
	}
	assert.Equal(t, 1, len(finals))
	assert.Equal(t, []int{22, 33}, finals[chunk])
	assert.Equal(t, []int{22, 33}, ends[chunk])
}

func TestChainErrRecovery(t *testing.T) {
	someErr := errors.New("some err")
	okCount := 0
	recovered := []error{}
	reset := func() {
		okCount = 0
		recovered = recovered[:0]
	}
	okp := procs.InplaceFunc(func(*scat.Chunk) error {
		okCount++
		return nil
	})
	errp := procs.InplaceFunc(func(*scat.Chunk) error {
		return someErr
	})
	errpNoChunk := procs.ProcFunc(func(*scat.Chunk) <-chan procs.Res {
		ch := make(chan procs.Res, 1)
		defer close(ch)
		ch <- procs.Res{Err: someErr}
		return ch
	})
	recover := errProcFunc(func(c *scat.Chunk, err error) <-chan procs.Res {
		ch := make(chan procs.Res, 1)
		defer close(ch)
		ch <- procs.Res{Chunk: c}
		recovered = append(recovered, err)
		return ch
	})
	recoverFail := errProcFunc(func(c *scat.Chunk, err error) <-chan procs.Res {
		ch := make(chan procs.Res, 1)
		defer close(ch)
		ch <- procs.Res{Chunk: c, Err: err}
		return ch
	})

	// no recovery
	reset()
	chain := procs.Chain{errp, okp}
	err := getErr(t, chain.Process(scat.NewChunk(0, nil)))
	assert.Equal(t, someErr, err)
	assert.Equal(t, 0, okCount)
	assert.Equal(t, []error{}, recovered)

	// recovery
	reset()
	chain = procs.Chain{errp, okp, recover, okp}
	err = getErr(t, chain.Process(scat.NewChunk(0, nil)))
	assert.NoError(t, err)
	assert.Equal(t, 1, okCount)
	assert.Equal(t, []error{someErr}, recovered)

	// failed recovery
	reset()
	chain = procs.Chain{errp, okp, recoverFail, okp}
	err = getErr(t, chain.Process(scat.NewChunk(0, nil)))
	assert.Equal(t, someErr, err)
	assert.Equal(t, 0, okCount)
	assert.Equal(t, []error{}, recovered)

	// impossible recovery: err without chunk
	reset()
	chain = procs.Chain{errpNoChunk, okp, recover, okp}
	err = getErr(t, chain.Process(scat.NewChunk(0, nil)))
	assert.Equal(t, someErr, err)
	assert.Equal(t, 0, okCount)
	assert.Equal(t, []error{}, recovered)
}

func TestChainFinish(t *testing.T) {
	testutil.TestFinishErrForward(t, func(proc procs.Proc) testutil.Finisher {
		return procs.Chain{procs.Nop, proc}
	})
}

type enderProc struct {
	proc    procs.Proc
	onFinal func(*scat.Chunk, *scat.Chunk) error
	onEnd   func(*scat.Chunk) error
}

type ender interface {
	procs.Proc
	procs.EndProc
}

var _ ender = enderProc{}

func (e enderProc) Process(c *scat.Chunk) <-chan procs.Res {
	return e.proc.Process(c)
}

func (e enderProc) Finish() error {
	return e.proc.Finish()
}

func (e enderProc) ProcessFinal(c, final *scat.Chunk) error {
	return e.onFinal(c, final)
}

func (e enderProc) ProcessEnd(c *scat.Chunk) error {
	return e.onEnd(c)
}

type recoverProc interface {
	procs.Proc
	procs.ErrProc
}

type errProcFunc func(*scat.Chunk, error) <-chan procs.Res

var _ recoverProc = errProcFunc(func(*scat.Chunk, error) <-chan procs.Res {
	return nil
})

func (fn errProcFunc) Process(c *scat.Chunk) <-chan procs.Res {
	return procs.Nop.Process(c)
}

func (fn errProcFunc) ProcessErr(c *scat.Chunk, err error) <-chan procs.Res {
	return fn(c, err)
}

func (fn errProcFunc) Finish() error {
	return nil
}
