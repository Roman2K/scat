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

	called := []string{}
	testProc := func(id string) aprocs.Proc {
		return aprocs.InplaceProcFunc(func(*ss.Chunk) error {
			called = append(called, id)
			return nil
		})
	}

	copiers := []cpprocs.Copier{
		{
			Id:     "a",
			Lister: testutil.SliceLister{hash1},
			Proc:   testProc("a"),
		}, {
			Id:     "b",
			Lister: testutil.SliceLister{hash1, hash2},
			Proc:   testProc("b"),
		}, {
			Id:     "c",
			Lister: testutil.SliceLister{},
			Proc:   testProc("c"),
		},
	}

	var mc aprocs.DynProcer
	resetMc := func() {
		var err error
		mc, err = New(min, copiers)
		assert.NoError(t, err)
	}
	resetCalled := func() {
		called = called[:0]
	}
	reset := func() {
		resetMc()
		resetCalled()
	}

	testProcsForHash := func(h checksum.Hash, expectedCalls []string) {
		c := &ss.Chunk{Hash: h}
		procs, err := mc.Procs(c)
		assert.NoError(t, err)
		assert.Equal(t, len(expectedCalls)+1, len(procs))
		chunks, err := processByAll(c, procs)
		assert.NoError(t, err)
		assert.Equal(t, []*ss.Chunk{c}, chunks)
		assert.Equal(t, expectedCalls, called)
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
	testProcsForHash(hash3, []string{"c", "b"})

	reset()
	rand2 = func() int { return 1 }
	testProcsForHash(hash2, []string{"a"})
	resetCalled()
	testProcsForHash(hash2, []string{})
	resetCalled()
	testProcsForHash(hash2, []string{})
}

func TestMinCopiesFinish(t *testing.T) {
	copiers := []cpprocs.Copier{
		{
			Id:     "",
			Lister: testutil.SliceLister{},
			Proc:   testutil.FinishErrProc{Err: nil},
		},
	}
	mc, err := New(2, copiers)
	assert.NoError(t, err)
	err = mc.Finish()
	assert.NoError(t, err)

	someErr := errors.New("some err")
	copiers = []cpprocs.Copier{
		{
			Id:     "",
			Lister: testutil.SliceLister{},
			Proc:   testutil.FinishErrProc{Err: someErr},
		},
	}
	mc, err = New(2, copiers)
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
