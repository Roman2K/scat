package mincopies

import (
	"errors"
	"sort"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/checksum"
	"scat/procs"
	"scat/stores"
	"scat/stores/quota"
	"scat/testutil"
)

var byId = testutil.SortCopiersByIdString

func TestMinCopies(t *testing.T) {
	const min = 2

	shuffleOrig := shuffle
	defer func() {
		shuffle = shuffleOrig
	}()

	hash1 := checksum.SumBytes([]byte("hash1"))
	hash2 := checksum.SumBytes([]byte("hash2"))
	hash3 := checksum.SumBytes([]byte("hash3"))
	hash4 := checksum.SumBytes([]byte("hash4"))

	called := []string{}
	errs := map[string]error{}
	testProc := func(id string) procs.Proc {
		return procs.InplaceFunc(func(*scat.Chunk) error {
			called = append(called, id)
			return errs[id]
		})
	}

	newQman := func() (qman *quota.Man) {
		qman = quota.NewMan()
		qman.AddRes(stores.NewCopier("a",
			stores.SliceLister{{Hash: hash1}},
			testProc("a"),
		))
		qman.AddRes(stores.NewCopier("b",
			stores.SliceLister{{Hash: hash1}, {Hash: hash2}},
			testProc("b"),
		))
		qman.AddRes(stores.NewCopier("c",
			stores.SliceLister{},
			testProc("c"),
		))
		return
	}

	var mc procs.DynProcer
	resetMc := func() {
		var err error
		mc, err = New(min, newQman())
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

	testProcsForHash := func(h checksum.Hash, expectedCalls []string) error {
		c := chunkWithHash(h)
		procs, err := mc.Procs(c)
		assert.NoError(t, err)
		chunks, err := processByAll(c, procs)
		assert.Equal(t, 1, len(chunks))
		assert.Equal(t, []*scat.Chunk{c}, chunks)
		assert.Equal(t, expectedCalls, called)
		return err
	}
	testProcsForHashNoErr := func(h checksum.Hash, expectedCalls []string) {
		err := testProcsForHash(h, expectedCalls)
		assert.NoError(t, err)
	}

	reset()
	testProcsForHashNoErr(hash1, []string{})

	reset()
	shuffle = byId
	testProcsForHashNoErr(hash2, []string{"a"})
	resetCalled()
	testProcsForHashNoErr(hash2, []string{})

	reset()
	shuffle = reverse
	testProcsForHashNoErr(hash2, []string{"c"})
	resetCalled()
	testProcsForHashNoErr(hash2, []string{})

	reset()
	shuffle = byId
	testProcsForHashNoErr(hash3, []string{"a", "b"})
	resetCalled()
	testProcsForHashNoErr(hash3, []string{})

	reset()
	shuffle = reverse
	testProcsForHashNoErr(hash3, []string{"c", "b"})
	resetCalled()
	testProcsForHashNoErr(hash3, []string{})

	// failover: OK
	reset()
	shuffle = byId
	someErr := errors.New("some err")
	errs["a"] = someErr
	testProcsForHashNoErr(hash3, []string{"a", "c", "b"})
	resetCalled()
	testProcsForHashNoErr(hash3, []string{})
	resetCalled()
	testProcsForHashNoErr(hash4, []string{"b", "c"})

	// failover: all KO
	reset()
	shuffle = byId
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	errs["a"] = err1
	errs["c"] = err2
	err := testProcsForHash(hash3, []string{"a", "c", "b"})
	assert.Equal(t, err2, err)
	resetCalled()
	_, err = mc.Procs(chunkWithHash(hash3))
	assert.Equal(t, "missing copiers to meet min requirement:"+
		" min=2 copies=1 missing=1 avail=0",
		err.Error(),
	)
}

func TestMinCopiesNegativeMissing(t *testing.T) {
	called := []string{}
	testProc := func(id string) procs.Proc {
		return procs.InplaceFunc(func(*scat.Chunk) error {
			called = append(called, id)
			return nil
		})
	}

	hash1 := checksum.SumBytes([]byte("hash1"))
	qman := quota.NewMan()
	qman.AddRes(stores.NewCopier("a",
		stores.SliceLister{{Hash: hash1}},
		testProc("a"),
	))
	qman.AddRes(stores.NewCopier("b",
		stores.SliceLister{{Hash: hash1}},
		testProc("b"),
	))
	mc, err := New(1, qman)
	assert.NoError(t, err)

	procs, err := mc.Procs(chunkWithHash(hash1))
	assert.NoError(t, err)
	_, err = processByAll(chunkWithHash(hash1), procs)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, called)
}

func TestFinish(t *testing.T) {
	qman := quota.NewMan()
	qman.AddRes(stores.NewCopier(nil,
		stores.SliceLister{},
		testutil.FinishErrProc{Err: nil},
	))
	mc, err := New(2, qman)
	assert.NoError(t, err)
	err = mc.Finish()
	assert.NoError(t, err)
}

func TestFinishError(t *testing.T) {
	someErr := errors.New("some err")
	qman := quota.NewMan()
	qman.AddRes(stores.NewCopier(nil,
		stores.SliceLister{},
		testutil.FinishErrProc{Err: someErr},
	))
	mc, err := New(2, qman)
	assert.NoError(t, err)
	err = mc.Finish()
	assert.Equal(t, someErr, err)
}

func processByAll(c *scat.Chunk, procs []procs.Proc) (
	all []*scat.Chunk, err error,
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
	s := []stores.Copier{
		stores.NewCopier("a", nil, nil),
		stores.NewCopier("b", nil, nil),
		stores.NewCopier("c", nil, nil),
	}
	ids := ids(shuffle(s))
	sort.Strings(ids)
	assert.Equal(t, []string{"a", "b", "c"}, ids)
}

func reverse(s []stores.Copier) (res []stores.Copier) {
	s = byId(s)
	n := len(s)
	res = make([]stores.Copier, n)
	for i := 0; i < n; i++ {
		res[i] = s[n-i-1]
	}
	return
}

func TestReverse(t *testing.T) {
	s := []stores.Copier{
		stores.NewCopier("a", nil, nil),
		stores.NewCopier("b", nil, nil),
		stores.NewCopier("c", nil, nil),
	}
	assert.Equal(t, []string{"c", "b", "a"}, ids(reverse(s)))
}

func ids(s []stores.Copier) (ids []string) {
	for _, c := range s {
		ids = append(ids, c.Id().(string))
	}
	return
}

func chunkWithHash(h checksum.Hash) (c *scat.Chunk) {
	c = scat.NewChunk(0, nil)
	c.SetHash(h)
	return
}
