package cpprocs_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"

	"secsplit/checksum"
	"secsplit/cpprocs"
)

func TestCatProcCmd(t *testing.T) {
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
	entries, err := cat.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(entries))

	// NewProcCmd()
	hash := checksum.Sum([]byte(hashData))
	cmd, err := cat.NewProcCmd(hash)
	assert.NoError(t, err)
	cmd.Stdin = strings.NewReader(data)
	err = cmd.Run()
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

func TestCatUnprocCmd(t *testing.T) {
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
	cmd, err := cat.NewUnprocCmd(hash)
	assert.NoError(t, err)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	err = cmd.Run()
	assert.NoError(t, err)
	assert.Equal(t, data, buf.String())
}
