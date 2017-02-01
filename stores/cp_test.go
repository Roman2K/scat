package stores_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/procs"
	"github.com/Roman2K/scat/stores"
	"github.com/Roman2K/scat/testutil"
)

func TestCp(t *testing.T) {
	dirStoreTest(func(dir stores.Dir) stores.Store {
		return stores.Cp(dir)
	}).run(t)
}

type dirStoreTest func(stores.Dir) stores.Store

func (test dirStoreTest) run(t *testing.T) {
	test.testReadWrite(t)
	test.testInvalidDir(t)
	test.testMissingData(t)
	test.testLs(t)
	test.testLsMissingDir(t)
}

func (test dirStoreTest) testReadWrite(t *testing.T) {
	const (
		data = "abc"
	)
	var (
		hash = testutil.Hash1.Hash
		hex  = testutil.Hash1.Hex
	)

	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	testPart := func(part stores.StrPart, expectedPath string) {
		store := test(stores.Dir{dir, part})

		// write
		c := scat.NewChunk(0, scat.BytesData(data))
		c.SetHash(hash)
		chunks, err := testutil.ReadChunks(store.Proc().Process(c))
		assert.NoError(t, err)
		assert.Equal(t, []*scat.Chunk{c}, chunks)
		b, err := ioutil.ReadFile(expectedPath)
		assert.NoError(t, err)
		assert.Equal(t, data, string(b))

		// read
		c = scat.NewChunk(0, nil)
		c.SetHash(hash)
		chunks, err = testutil.ReadChunks(store.Unproc().Process(c))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(chunks))
		b, err = chunks[0].Data().Bytes()
		assert.NoError(t, err)
		assert.Equal(t, data, string(b))
	}

	testPart(nil, filepath.Join(dir, hex))
	testPart(stores.StrPart{2}, filepath.Join(dir, hex[:2], hex))
	testPart(stores.StrPart{2, 3}, filepath.Join(dir, hex[:2], hex[2:5], hex))
}

func (test dirStoreTest) testInvalidDir(t *testing.T) {
	store := test(stores.Dir{Path: "/dev/null"})
	c := scat.NewChunk(0, nil)
	_, err := testutil.ReadChunks(store.Proc().Process(c))
	assert.Error(t, err)
	if exit, ok := err.(*exec.ExitError); ok {
		assert.Regexp(t, "Not a directory", string(exit.Stderr))
	} else {
		assert.Regexp(t, "not a directory", err.Error())
	}
}

func (test dirStoreTest) testMissingData(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	store := test(stores.Dir{Path: dir + "/missing"})
	c := scat.NewChunk(0, nil)
	_, err = testutil.ReadChunks(store.Unproc().Process(c))
	missErr, ok := err.(procs.MissingDataError)
	assert.True(t, ok)
	assert.Error(t, missErr.Err)
}

func (test dirStoreTest) testLs(t *testing.T) {
	const (
		fileMode = 0644
		dirMode  = 0755
	)
	var (
		hash  = testutil.Hashes[0].Hash
		hex   = testutil.Hashes[0].Hex
		hash2 = testutil.Hashes[1].Hash
		hex2  = testutil.Hashes[1].Hex
	)

	// depth=0
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	store := test(stores.Dir{Path: dir})

	// depth=0 files=0 chunkFiles=0
	ls, err := store.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(ls))

	// depth=0 files=1 chunkFiles=0
	err = ioutil.WriteFile(filepath.Join(dir, "xxx"), nil, fileMode)
	assert.NoError(t, err)
	ls, err = store.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(ls))

	// depth=0 files=2 chunkFiles=1
	err = ioutil.WriteFile(filepath.Join(dir, hex), []byte("x"), fileMode)
	assert.NoError(t, err)
	ls, err = store.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ls))
	assert.Equal(t, hash, ls[0].Hash)
	assert.Equal(t, int64(1), ls[0].Size)

	// depth=1
	dir, err = ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	store = test(stores.Dir{dir, stores.StrPart{1}})

	// depth=1 files=0 chunkFiles=0
	ls, err = store.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(ls))

	// depth=1 files=1 chunkFiles=0
	err = ioutil.WriteFile(filepath.Join(dir, hex), nil, fileMode)
	assert.NoError(t, err)
	ls, err = store.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(ls))

	// depth=1 files=1 chunkFiles=1
	path := filepath.Join(dir, hex[:1], hex)
	err = os.Mkdir(filepath.Dir(path), dirMode)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path, []byte("a"), fileMode)
	assert.NoError(t, err)
	ls, err = store.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ls))
	assert.Equal(t, hash, ls[0].Hash)
	assert.Equal(t, int64(1), ls[0].Size)

	// depth=1 files=2 chunkFiles=2
	path = filepath.Join(dir, hex2[:1], hex2)
	err = os.MkdirAll(filepath.Dir(path), dirMode)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path, []byte("a"), fileMode)
	assert.NoError(t, err)
	ls, err = store.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(ls))
	getHex := func(i int) string {
		return fmt.Sprintf("%x", ls[i].Hash)
	}
	sort.Slice(ls, func(i, j int) bool {
		return getHex(i) < getHex(j)
	})
	assert.Equal(t, hash, ls[0].Hash)
	assert.Equal(t, hash2, ls[1].Hash)

	// depth=1 files=1 dirs=1 chunkFiles=1
	err = os.Remove(path)
	assert.NoError(t, err)
	err = os.Mkdir(path, dirMode)
	assert.NoError(t, err)
	ls, err = store.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ls))
	assert.Equal(t, hash, ls[0].Hash)
}

func (test dirStoreTest) testLsMissingDir(t *testing.T) {
	store := test(stores.Dir{Path: "/dev/nullxxx"})
	_, err := store.Ls()
	assert.Error(t, err)
	if exit, ok := err.(*exec.ExitError); ok {
		assert.Regexp(t, "No such file", string(exit.Stderr))
	} else {
		assert.True(t, os.IsNotExist(err))
	}
}
