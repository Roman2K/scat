package procs_test

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"scat"
	"scat/procs"
	"scat/testutil"
)

func TestCmdInFunc(t *testing.T) {
	const data = "xxx"
	buf := &bytes.Buffer{}
	cmdp := procs.CmdInFunc(func(scat.Chunk) (*exec.Cmd, error) {
		cmd := exec.Command("cat")
		cmd.Stdout = buf
		return cmd, nil
	})
	c := scat.NewChunk(0, []byte(data))
	ch := cmdp.Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.NoError(t, err)
	assert.Equal(t, []scat.Chunk{c}, chunks)
	assert.Equal(t, data, buf.String())
}

func TestCmdInFuncError(t *testing.T) {
	cmdp := procs.CmdInFunc(func(scat.Chunk) (*exec.Cmd, error) {
		return exec.Command("/dev/null"), nil
	})
	c := scat.NewChunk(0, nil)
	ch := cmdp.Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.Equal(t, []scat.Chunk{c}, chunks)
	assert.Equal(t, "fork/exec /dev/null: permission denied", err.Error())
}

func TestCmdOutFunc(t *testing.T) {
	const output = "ok"
	cmdp := procs.CmdOutFunc(func(scat.Chunk) (*exec.Cmd, error) {
		return exec.Command("echo", "-n", output), nil
	})
	c := scat.NewChunk(0, nil)
	ch := cmdp.Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))
	assert.Equal(t, output, string(chunks[0].Data()))
}

func TestCmdOutFuncError(t *testing.T) {
	cmdp := procs.CmdOutFunc(func(scat.Chunk) (*exec.Cmd, error) {
		return exec.Command("/dev/null"), nil
	})
	c := scat.NewChunk(0, nil)
	ch := cmdp.Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.Equal(t, []scat.Chunk{c}, chunks)
	assert.Equal(t, "fork/exec /dev/null: permission denied", err.Error())
	assert.Nil(t, c.Data())
}
