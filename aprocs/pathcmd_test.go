package aprocs_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
	"secsplit/testutil"
	"secsplit/tmpdedup"
)

func TestPathCmdIn(t *testing.T) {
	const (
		data     = "xxx"
		hashData = "x"
		hashStr  = "2d711642b726b04401627ca9fbac32f5c8530fb1903cc4db02258717921a4881"
	)
	fdata := &bytes.Buffer{}
	paths := []string{}
	newCmd := func(_ *ss.Chunk, path string) (*exec.Cmd, error) {
		paths = append(paths, path)
		cmd := exec.Command("cat", path)
		cmd.Stdout = fdata
		return cmd, nil
	}
	cmdp, err := aprocs.NewPathCmdIn("", newCmd)
	assert.NoError(t, err)
	c := &ss.Chunk{
		Hash: checksum.Sum([]byte(hashData)),
		Data: []byte(data),
	}
	chunks, err := testutil.ReadChunks(cmdp.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
	assert.Equal(t, data, string(chunks[0].Data))
	assert.Equal(t, 1, len(paths))
	assert.Equal(t, hashStr, filepath.Base(paths[0]))
	assert.Equal(t, data, fdata.String())

	type tmpManer interface {
		TmpMan() *tmpdedup.Man
	}

	// tmp cleanup
	cmdp.(tmpManer).TmpMan().Wait()
	_, err = os.Stat(paths[0])
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	// Finish()
	dir := filepath.Dir(paths[0])
	_, err = os.Stat(dir)
	assert.NoError(t, err)
	err = cmdp.Finish()
	assert.NoError(t, err)
	_, err = os.Stat(dir)
	assert.True(t, os.IsNotExist(err))

	// idempotence
	err = cmdp.Finish()
	assert.NoError(t, err)
}
