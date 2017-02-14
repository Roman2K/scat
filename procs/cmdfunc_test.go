package procs_test

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/testutil"
)

func TestCmdFunc(t *testing.T) {
	const data = "xxx"
	cmdp := procs.CmdFunc(func(*scat.Chunk) (*exec.Cmd, error) {
		return exec.Command("cat"), nil
	})
	c := scat.NewChunk(0, scat.BytesData(data))
	chunks, err := testutil.ReadChunks(cmdp.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)
	b, err := c.Data().Bytes()
	assert.NoError(t, err)
	assert.Equal(t, data, string(b))
}

func TestCmdFuncError(t *testing.T) {
	testCmdFuncError(t, func(fn procs.CmdFunc) procs.Proc {
		return fn
	})
}

func TestCmdInFunc(t *testing.T) {
	const data = "xxx"
	buf := &bytes.Buffer{}
	cmdp := procs.CmdInFunc(func(*scat.Chunk) (*exec.Cmd, error) {
		cmd := exec.Command("sh", "-c", "echo ok && cat")
		cmd.Stdout = buf
		return cmd, nil
	})
	c := scat.NewChunk(0, scat.BytesData(data))
	chunks, err := testutil.ReadChunks(cmdp.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)
	assert.Equal(t, "ok\n"+data, buf.String())
	b, err := c.Data().Bytes()
	assert.NoError(t, err)
	assert.Equal(t, data, string(b))
}

func TestCmdInFuncError(t *testing.T) {
	testCmdFuncError(t, func(fn procs.CmdFunc) procs.Proc {
		return procs.CmdInFunc(fn)
	})
}

func TestCmdOutFunc(t *testing.T) {
	const output = "ok"
	cmdp := procs.CmdOutFunc(func(*scat.Chunk) (*exec.Cmd, error) {
		return exec.Command("echo", "-n", output), nil
	})
	c := scat.NewChunk(0, nil)
	chunks, err := testutil.ReadChunks(cmdp.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))
	b, err := chunks[0].Data().Bytes()
	assert.NoError(t, err)
	assert.Equal(t, output, string(b))
}

func TestCmdOutFuncError(t *testing.T) {
	testCmdFuncError(t, func(fn procs.CmdFunc) procs.Proc {
		return procs.CmdOutFunc(fn)
	})
}

func TestCmdOutFuncCustomStderr(t *testing.T) {
	const errOut = "xxx"
	errBuf := &bytes.Buffer{}
	fn := procs.CmdFunc(func(*scat.Chunk) (*exec.Cmd, error) {
		cmd := exec.Command("bash", "-c", fmt.Sprintf(
			`echo -n %q >&2; exit 1`, errOut,
		))
		cmd.Stderr = errBuf
		return cmd, nil
	})
	c := scat.NewChunk(0, nil)
	_, err := testutil.ReadChunks(fn.Process(c))
	exit, ok := err.(*exec.ExitError)
	assert.True(t, ok)
	assert.Equal(t, 0, len(exit.Stderr))
	assert.Equal(t, errOut, errBuf.String())
}

func testCmdFuncError(t *testing.T, getProc func(procs.CmdFunc) procs.Proc) {
	const errOut = "xxx"
	fn := procs.CmdFunc(func(*scat.Chunk) (*exec.Cmd, error) {
		cmd := exec.Command("bash", "-c", fmt.Sprintf(
			`echo ok; echo -n %q >&2; exit 1`, errOut,
		))
		return cmd, nil
	})
	cmdp := getProc(fn)
	c := scat.NewChunk(0, nil)
	chunks, err := testutil.ReadChunks(cmdp.Process(c))
	exit, ok := err.(*exec.ExitError)
	assert.True(t, ok)
	assert.Equal(t, []*scat.Chunk{c}, chunks)
	assert.Equal(t, errOut, string(exit.Stderr))
}
