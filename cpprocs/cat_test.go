package cpprocs_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/checksum"
	"secsplit/cpprocs"
	"secsplit/testutil"
)

func TestCatProc(t *testing.T) {
	const (
		hashData = "x"
		hashStr  = "2d711642b726b04401627ca9fbac32f5c8530fb1903cc4db02258717921a4881"
		data     = "xxx"
	)
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	cat := cpprocs.NewCat(dir)

	// Ls() empty
	entries, err := cat.LsProc().Ls()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(entries))

	// LsProc()
	hash := checksum.Sum([]byte(hashData))
	_, err = testutil.ReadChunks(cat.LsProc().Process(&ss.Chunk{
		Hash: hash,
		Data: []byte(data),
	}))
	assert.NoError(t, err)
	path := filepath.Join(dir, hashStr)
	fdata, err := ioutil.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, data, string(fdata))

	// Ls()
	entries, err = cat.LsProc().Ls()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, hash, entries[0].Hash)
	assert.Equal(t, int64(len(data)), entries[0].Size)
}

func TestCatUnproc(t *testing.T) {
	const (
		hashData = "y"
		hashStr  = "a1fce4363854ff888cff4b8e7875d600c2682390412a8cf79b37d0b11148b0fa"
		data     = "xxx"
	)
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	cat := cpprocs.NewCat(dir)

	path := filepath.Join(dir, hashStr)
	err = ioutil.WriteFile(path, []byte(data), 0644)
	assert.NoError(t, err)

	hash := checksum.Sum([]byte(hashData))
	chunks, err := testutil.ReadChunks(cat.LsUnproc().Process(&ss.Chunk{
		Hash: hash,
	}))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))
	assert.Equal(t, data, string(chunks[0].Data))
}
