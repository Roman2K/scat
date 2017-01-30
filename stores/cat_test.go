package stores_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/checksum"
	"scat/stores"
	"scat/testutil"
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
	cat := stores.NewCat(dir)

	// Ls() empty
	entries, err := cat.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(entries))

	// Proc()
	hash := checksum.SumBytes([]byte(hashData))
	c := scat.NewChunk(0, scat.BytesData(data))
	c.SetHash(hash)
	_, err = testutil.ReadChunks(cat.Proc().Process(c))
	assert.NoError(t, err)
	path := filepath.Join(dir, hashStr)
	fdata, err := ioutil.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, data, string(fdata))

	// Ls()
	entries, err = cat.Ls()
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
	cat := stores.NewCat(dir)

	path := filepath.Join(dir, hashStr)
	err = ioutil.WriteFile(path, []byte(data), 0644)
	assert.NoError(t, err)

	hash := checksum.SumBytes([]byte(hashData))
	c := scat.NewChunk(0, nil)
	c.SetHash(hash)
	chunks, err := testutil.ReadChunks(cat.Unproc().Process(c))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))
	b, err := chunks[0].Data().Bytes()
	assert.NoError(t, err)
	assert.Equal(t, data, string(b))
}
