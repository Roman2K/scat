package cpprocs_test

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	ss "secsplit"
	"secsplit/checksum"
	"secsplit/cpprocs"
	"secsplit/testutil"
)

func TestCommandLsProc(t *testing.T) {
	const data = "xxx"
	buf := &bytes.Buffer{}
	spawner := testSpawner{
		newProcCmd: func(checksum.Hash) (*exec.Cmd, error) {
			cmd := exec.Command("cat")
			cmd.Stdout = buf
			return cmd, nil
		},
	}
	cmd := cpprocs.NewCommand(spawner)
	c := &ss.Chunk{Data: []byte(data)}
	ch := cmd.LsProc().Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.NoError(t, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
	assert.Equal(t, data, buf.String())
}

func TestCommandLsProcError(t *testing.T) {
	spawner := testSpawner{
		newProcCmd: func(checksum.Hash) (*exec.Cmd, error) {
			return exec.Command("/dev/null"), nil
		},
	}
	cmd := cpprocs.NewCommand(spawner)
	c := &ss.Chunk{}
	ch := cmd.LsProc().Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
	assert.Equal(t, "fork/exec /dev/null: permission denied", err.Error())
}

func TestCommandLsUnproc(t *testing.T) {
	const output = "ok"
	spawner := testSpawner{
		newUnprocCmd: func(checksum.Hash) (*exec.Cmd, error) {
			return exec.Command("echo", "-n", output), nil
		},
	}
	cmd := cpprocs.NewCommand(spawner)
	c := &ss.Chunk{}
	ch := cmd.LsUnproc().Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.NoError(t, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
	assert.Equal(t, output, string(chunks[0].Data))
}

func TestCommandLsUnprocError(t *testing.T) {
	spawner := testSpawner{
		newUnprocCmd: func(checksum.Hash) (*exec.Cmd, error) {
			return exec.Command("/dev/null"), nil
		},
	}
	cmd := cpprocs.NewCommand(spawner)
	c := &ss.Chunk{}
	ch := cmd.LsUnproc().Process(c)
	chunks, err := testutil.ReadChunks(ch)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
	assert.Equal(t, "fork/exec /dev/null: permission denied", err.Error())
	assert.Nil(t, c.Data)
}

type testSpawner struct {
	newProcCmd   func(checksum.Hash) (*exec.Cmd, error)
	newUnprocCmd func(checksum.Hash) (*exec.Cmd, error)
}

func (ts testSpawner) Ls() ([]cpprocs.LsEntry, error) {
	panic("Ls() not implemented")
}

func (ts testSpawner) NewProcCmd(h checksum.Hash) (*exec.Cmd, error) {
	if ts.newProcCmd == nil {
		panic("newProcCmd not set")
	}
	return ts.newProcCmd(h)
}

func (ts testSpawner) NewUnprocCmd(h checksum.Hash) (*exec.Cmd, error) {
	if ts.newUnprocCmd == nil {
		panic("newUnprocCmd not set")
	}
	return ts.newUnprocCmd(h)
}
