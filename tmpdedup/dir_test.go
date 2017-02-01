package tmpdedup_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Roman2K/scat/tmpdedup"
	assert "github.com/stretchr/testify/require"
)

func TestDir(t *testing.T) {
	const (
		filename = "xxx"
		data     = "yyy"
	)

	parent, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(parent)

	dir, err := tmpdedup.TempDir(parent)
	assert.NoError(t, err)
	defer dir.Finish()

	path, wg, err := dir.Get(filename, func(path string) error {
		return ioutil.WriteFile(path, []byte(data), 0600)
	})
	assert.NoError(t, err)
	assert.Equal(t, filename, filepath.Base(path))
	assert.Equal(t, parent, filepath.Dir(filepath.Dir(path)))

	fdata, err := ioutil.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, data, string(fdata))

	assertNotExists := func(path string) {
		_, err := os.Stat(path)
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	}

	wg.Done()
	dir.TmpMan().Wait()
	assertNotExists(path)

	// Finish()
	err = dir.Finish()
	assert.NoError(t, err)
	assertNotExists(filepath.Dir(path))
	_, err = os.Stat(parent)
	assert.NoError(t, err)

	// idempotence
	err = dir.Finish()
	assert.NoError(t, err)
}

func TestDirError(t *testing.T) {
	dir, err := tmpdedup.TempDir("")
	assert.NoError(t, err)
	defer dir.Finish()

	someErr := errors.New("some err")
	_, _, err = dir.Get("xxx", func(string) error {
		return someErr
	})
	assert.Equal(t, someErr, err)
}
