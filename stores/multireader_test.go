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

func TestMultiReader(t *testing.T) {
	var (
		hash = testutil.Hash1.Hash
		hex  = testutil.Hash1.Hex
	)

	dir1, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	dir2, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir1)
	defer os.RemoveAll(dir2)

	cp1 := stores.Cp{Dir: dir1}
	cp2 := stores.Cp{Dir: dir2}
	copiers := []stores.Copier{
		stores.NewCopier("cp1", cp1, cp1.Unproc()),
		stores.NewCopier("cp2", cp2, cp2.Unproc()),
	}

	c := scat.NewChunk(0, nil)
	c.SetHash(hash)

	// none available
	mrd, err := stores.NewMultiReader(copiers)
	assert.NoError(t, err)
	_, err = testutil.ReadChunks(mrd.Process(c))
	assert.Equal(t, stores.ErrMultiReaderNoneAvail, err)

	readData := func() string {
		chunks, err := testutil.ReadChunks(mrd.Process(c))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(chunks))
		b, err := chunks[0].Data().Bytes()
		assert.NoError(t, err)
		return string(b)
	}

	// on cp2
	err = ioutil.WriteFile(filepath.Join(dir2, hex), []byte("data2"), 0644)
	assert.NoError(t, err)
	mrd, err = stores.NewMultiReader(copiers)
	assert.NoError(t, err)
	assert.Equal(t, "data2", readData())

	// on cp2 and cp1
	err = ioutil.WriteFile(filepath.Join(dir1, hex), []byte("data1"), 0644)
	assert.NoError(t, err)
	mrd, err = stores.NewMultiReader(copiers)
	assert.NoError(t, err)
	assert.Equal(t, "data1", readData())
}
