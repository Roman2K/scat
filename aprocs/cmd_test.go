package aprocs_test

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/testutil"
)

func TestCmdInFunc(t *testing.T) {
	const data = "xxx"
	buf := &bytes.Buffer{}
	cmdp := aprocs.CmdInFunc(func(*ss.Chunk) (*exec.Cmd, error) {
		cmd := exec.Command("cat")
		cmd.Stdout = buf
		return cmd, nil
	})
	c := &ss.Chunk{Data: []byte(data)}
	ch := cmdp.Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.NoError(t, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
	assert.Equal(t, data, buf.String())
}

func TestCmdInFuncError(t *testing.T) {
	cmdp := aprocs.CmdInFunc(func(*ss.Chunk) (*exec.Cmd, error) {
		return exec.Command("/dev/null"), nil
	})
	c := &ss.Chunk{}
	ch := cmdp.Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
	assert.Equal(t, "fork/exec /dev/null: permission denied", err.Error())
}

func TestCmdOutFunc(t *testing.T) {
	const output = "ok"
	cmdp := aprocs.CmdOutFunc(func(*ss.Chunk) (*exec.Cmd, error) {
		return exec.Command("echo", "-n", output), nil
	})
	c := &ss.Chunk{}
	ch := cmdp.Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.NoError(t, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
	assert.Equal(t, output, string(chunks[0].Data))
}

func TestCmdOutFuncError(t *testing.T) {
	cmdp := aprocs.CmdOutFunc(func(*ss.Chunk) (*exec.Cmd, error) {
		return exec.Command("/dev/null"), nil
	})
	c := &ss.Chunk{}
	ch := cmdp.Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
	assert.Equal(t, "fork/exec /dev/null: permission denied", err.Error())
	assert.Nil(t, c.Data)
}
