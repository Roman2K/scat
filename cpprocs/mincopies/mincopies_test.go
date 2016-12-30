package mincopies

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
	"secsplit/cpprocs"
	"secsplit/testutil"
)

func TestMinCopies(t *testing.T) {
	const min = 2

	hash1 := checksum.Sum([]byte("hash1"))
	hash2 := checksum.Sum([]byte("hash2"))
	hash3 := checksum.Sum([]byte("hash3"))

	cpps := []cpprocs.Proc{
		&testCpProc{
			id:     "a",
			hashes: []checksum.Hash{hash1},
		},
		&testCpProc{
			id:     "b",
			hashes: []checksum.Hash{hash1, hash2},
		},
		&testCpProc{
			id:     "c",
			hashes: []checksum.Hash{},
		},
	}

	calledProcs := func() []string {
		ids := []string{}
		for _, cpp := range cpps {
			if cpp.(*testCpProc).called {
				ids = append(ids, cpp.Id().(string))
			}
		}
		return ids
	}

	var mc aprocs.DynProcer
	resetMc := func() {
		var err error
		mc, err = New(min, cpps)
		assert.NoError(t, err)
	}
	resetProcs := func() {
		for _, cpp := range cpps {
			cpp.(*testCpProc).called = false
		}
	}
	reset := func() {
		resetMc()
		resetProcs()
	}

	testProcsForHash := func(h checksum.Hash, expectedCalls []string) {
		c := &ss.Chunk{Hash: h}
		procs, err := mc.Procs(c)
		assert.NoError(t, err)
		assert.Equal(t, len(expectedCalls)+1, len(procs))
		chunks, err := processByAll(c, procs)
		assert.NoError(t, err)
		assert.Equal(t, []*ss.Chunk{c}, chunks)
		assert.Equal(t, expectedCalls, calledProcs())
	}

	reset()
	testProcsForHash(hash1, []string{})

	reset()
	rand2 = func() int { return 1 }
	testProcsForHash(hash2, []string{"a"})

	reset()
	rand2 = func() int { return 0 }
	testProcsForHash(hash2, []string{"c"})

	reset()
	rand2 = func() int { return 1 }
	testProcsForHash(hash3, []string{"a", "b"})

	reset()
	rand2 = func() int { return 0 }
	testProcsForHash(hash3, []string{"b", "c"})

	reset()
	rand2 = func() int { return 1 }
	testProcsForHash(hash2, []string{"a"})
	resetProcs()
	testProcsForHash(hash2, []string{})
	resetProcs()
	testProcsForHash(hash2, []string{})
}

func TestMinCopiesFinish(t *testing.T) {
	a := &testCpProc{}
	mc, err := New(2, []cpprocs.Proc{a})
	assert.NoError(t, err)
	err = mc.Finish()
	assert.NoError(t, err)

	someErr := errors.New("some err")
	a = &testCpProc{finishErr: someErr}
	mc, err = New(2, []cpprocs.Proc{a})
	assert.NoError(t, err)
	err = mc.Finish()
	assert.Equal(t, someErr, err)
}

func processByAll(c *ss.Chunk, procs []aprocs.Proc) ([]*ss.Chunk, error) {
	all := []*ss.Chunk{}
	for _, proc := range procs {
		chunks, err := testutil.ReadChunks(proc.Process(c))
		if err != nil {
			return nil, err
		}
		all = append(all, chunks...)
	}
	return all, nil
}

type testCpProc struct {
	id        interface{}
	hashes    []checksum.Hash
	called    bool
	finishErr error
}

func (cpp *testCpProc) Id() interface{} {
	return cpp.id
}

func (cpp *testCpProc) Ls() ([]checksum.Hash, error) {
	return cpp.hashes, nil
}

func (cpp *testCpProc) Process(c *ss.Chunk) <-chan aprocs.Res {
	cpp.called = true
	ch := make(chan aprocs.Res)
	close(ch)
	return ch
}

func (cpp *testCpProc) Finish() error {
	return cpp.finishErr
}
