package mincopies

import (
	"errors"
	"sort"
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

	shuffleOrig := shuffle
	defer func() {
		shuffle = shuffleOrig
	}()

	hash1 := checksum.Sum([]byte("hash1"))
	hash2 := checksum.Sum([]byte("hash2"))
	hash3 := checksum.Sum([]byte("hash3"))
	hash4 := checksum.Sum([]byte("hash4"))

	called := []string{}
	errs := map[string]error{}
	testProc := func(id string) aprocs.Proc {
		return aprocs.InplaceProcFunc(func(*ss.Chunk) error {
			called = append(called, id)
			return errs[id]
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
	resetErrs := func() {
		for k := range errs {
			delete(errs, k)
		}
	}
	reset := func() {
		resetMc()
		resetCalled()
		resetErrs()
	}

	testProcsForHash := func(
		h checksum.Hash, expectedCalls []string, expectedErr error,
	) {
		c := &ss.Chunk{Hash: h}
		procs, err := mc.Procs(c)
		assert.NoError(t, err)
		chunks, err := processByAll(c, procs)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, len(chunks))
		assert.Equal(t, []*ss.Chunk{c}, chunks)
		assert.Equal(t, expectedCalls, called)
	}

	reset()
	testProcsForHash(hash1, []string{}, nil)

	reset()
	shuffle = identity
	testProcsForHash(hash2, []string{"a"}, nil)
	resetCalled()
	testProcsForHash(hash2, []string{}, nil)

	reset()
	shuffle = reverse
	testProcsForHash(hash2, []string{"c"}, nil)
	resetCalled()
	testProcsForHash(hash2, []string{}, nil)

	reset()
	shuffle = identity
	testProcsForHash(hash3, []string{"a", "b"}, nil)
	resetCalled()
	testProcsForHash(hash3, []string{}, nil)

	reset()
	shuffle = reverse
	testProcsForHash(hash3, []string{"c", "b"}, nil)
	resetCalled()
	testProcsForHash(hash3, []string{}, nil)

	// Failover: OK
	reset()
	shuffle = identity
	someErr := errors.New("some err")
	errs["a"] = someErr
	testProcsForHash(hash3, []string{"a", "c", "b"}, nil)
	resetCalled()
	testProcsForHash(hash3, []string{}, nil)
	resetCalled()
	testProcsForHash(hash4, []string{"b", "c"}, nil)

	// Failover: all KO
	reset()
	shuffle = identity
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	errs["a"] = err1
	errs["c"] = err2
	testProcsForHash(hash3, []string{"a", "c", "b"}, err2)
	resetCalled()
	testProcsForHash(hash3, []string{}, nil)
}

func TestFinish(t *testing.T) {
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
}

func TestFinishError(t *testing.T) {
	someErr := errors.New("some err")
	copiers := []cpprocs.Copier{
		{
			Id:     "",
			Lister: testutil.SliceLister{},
			Proc:   testutil.FinishErrProc{Err: someErr},
		},
	}
	mc, err := New(2, copiers)
	assert.NoError(t, err)
	err = mc.Finish()
	assert.Equal(t, someErr, err)
}

func processByAll(c *ss.Chunk, procs []aprocs.Proc) (
	all []*ss.Chunk, err error,
) {
	for _, proc := range procs {
		chunks, e := testutil.ReadChunks(proc.Process(c))
		if e != nil {
			if err == nil {
				err = e
			}
			continue
		}
		all = append(all, chunks...)
	}
	return
}

func TestShuffle(t *testing.T) {
	s := []cpprocs.Copier{
		{Id: 1},
		{Id: 2},
		{Id: 3},
	}
	ids := intIds(shuffle(s))
	sort.Ints(ids)
	assert.Equal(t, []int{1, 2, 3}, ids)
}

func identity(s []cpprocs.Copier) (res []cpprocs.Copier) {
	res = make([]cpprocs.Copier, len(s))
	copy(res, s)
	return
}

func reverse(s []cpprocs.Copier) (res []cpprocs.Copier) {
	n := len(s)
	res = make([]cpprocs.Copier, n)
	for i := 0; i < n; i++ {
		res[i] = s[n-i-1]
	}
	return
}

func TestReverseTest(t *testing.T) {
	s := []cpprocs.Copier{
		{Id: 1},
		{Id: 2},
		{Id: 3},
	}
	assert.Equal(t, []int{3, 2, 1}, intIds(reverse(s)))
}

func intIds(s []cpprocs.Copier) (ids []int) {
	for _, c := range s {
		ids = append(ids, c.Id.(int))
	}
	return
}
