package stores_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/stores"
	"scat/testutil"
)

func TestCp(t *testing.T) {
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

	testNest := func(nest stores.StrPart, expectedPath string) {
		cp := stores.Cp{Dir: dir, Nest: nest}

		// write
		c := scat.NewChunk(0, scat.BytesData(data))
		c.SetHash(hash)
		chunks, err := testutil.ReadChunks(cp.Proc().Process(c))
		assert.NoError(t, err)
		assert.Equal(t, []*scat.Chunk{c}, chunks)
		b, err := ioutil.ReadFile(expectedPath)
		assert.NoError(t, err)
		assert.Equal(t, data, string(b))

		// read
		c = scat.NewChunk(0, nil)
		c.SetHash(hash)
		chunks, err = testutil.ReadChunks(cp.Unproc().Process(c))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(chunks))
		b, err = chunks[0].Data().Bytes()
		assert.NoError(t, err)
		assert.Equal(t, data, string(b))
	}

	testNest(nil, filepath.Join(dir, hex))
	testNest(stores.StrPart{2}, filepath.Join(dir, hex[:2], hex))
	testNest(stores.StrPart{2, 3}, filepath.Join(dir, hex[:2], hex[2:5], hex))
}

func TestCpProcInvalidDir(t *testing.T) {
	cp := stores.Cp{Dir: "/dev/null"}
	c := scat.NewChunk(0, nil)
	_, err := testutil.ReadChunks(cp.Proc().Process(c))
	assert.Error(t, err)
	assert.Regexp(t, "not a directory", err.Error())
}

func TestCpUnprocMissingFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	cp := stores.Cp{Dir: dir + "/missing"}
	c := scat.NewChunk(0, nil)
	_, err = testutil.ReadChunks(cp.Unproc().Process(c))
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestCpLs(t *testing.T) {
	var (
		hash = testutil.Hash1.Hash
		hex  = testutil.Hash1.Hex
	)

	// depth=0
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	cp := stores.Cp{Dir: dir}

	// depth=0 files=0 chunkFiles=0
	ls, err := cp.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(ls))

	// depth=0 files=1 chunkFiles=0
	err = ioutil.WriteFile(filepath.Join(cp.Dir, "xxx"), nil, 0644)
	assert.NoError(t, err)
	ls, err = cp.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(ls))

	// depth=0 files=2 chunkFiles=1
	err = ioutil.WriteFile(filepath.Join(cp.Dir, hex), []byte("x"), 0644)
	assert.NoError(t, err)
	ls, err = cp.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ls))
	assert.Equal(t, hash, ls[0].Hash)
	assert.Equal(t, int64(1), ls[0].Size)

	// depth=1
	dir, err = ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	cp = stores.Cp{Dir: dir, Nest: stores.StrPart{1}}

	// depth=1 files=0 chunkFiles=0
	ls, err = cp.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(ls))

	// depth=1 files=1 chunkFiles=0
	err = ioutil.WriteFile(filepath.Join(cp.Dir, hex), nil, 0644)
	assert.NoError(t, err)
	ls, err = cp.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(ls))

	// depth=1 files=1 chunkFiles=1
	path := filepath.Join(cp.Dir, hex[:1], hex)
	err = os.Mkdir(filepath.Dir(path), 0755)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path, []byte("a"), 0644)
	assert.NoError(t, err)
	ls, err = cp.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ls))
	assert.Equal(t, hash, ls[0].Hash)
	assert.Equal(t, int64(1), ls[0].Size)
}
