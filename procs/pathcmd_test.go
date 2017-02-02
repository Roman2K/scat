package procs_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/checksum"
	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/testutil"
	"gitlab.com/Roman2K/scat/tmpdedup"
	assert "github.com/stretchr/testify/require"
)

func TestPathCmdIn(t *testing.T) {
	const (
		data     = "xxx"
		hashData = "x"
		hashStr  = "2d711642b726b04401627ca9fbac32f5c8530fb1903cc4db02258717921a4881"
	)
	fdata := &bytes.Buffer{}
	paths := []string{}
	newCmd := func(_ *scat.Chunk, path string) (*exec.Cmd, error) {
		paths = append(paths, path)
		cmd := exec.Command("cat", path)
		cmd.Stdout = fdata
		return cmd, nil
	}
	tmp, err := tmpdedup.TempDir("")
	assert.NoError(t, err)
	defer tmp.Finish()
	cmdp := procs.NewPathCmdIn(newCmd, tmp)
	c := scat.NewChunk(0, scat.BytesData(data))
	c.SetHash(checksum.SumBytes([]byte(hashData)))
	chunks, err := testutil.ReadChunks(cmdp.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)
	b, err := chunks[0].Data().Bytes()
	assert.NoError(t, err)
	assert.Equal(t, data, string(b))
	assert.Equal(t, 1, len(paths))
	assert.Equal(t, hashStr, filepath.Base(paths[0]))
	assert.Equal(t, data, fdata.String())

	// tmp cleanup
	tmp.TmpMan().Wait()
	_, err = os.Stat(paths[0])
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}
