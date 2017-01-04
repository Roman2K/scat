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

type testSpawner struct {
	newProcCmd func(checksum.Hash) (*exec.Cmd, error)
}

func (ts testSpawner) Ls() ([]cpprocs.LsEntry, error) {
	return nil, nil
}

func (ts testSpawner) NewProcCmd(h checksum.Hash) (*exec.Cmd, error) {
	return ts.newProcCmd(h)
}

func (ts testSpawner) NewUnprocCmd(h checksum.Hash) (*exec.Cmd, error) {
	panic("NewUnprocCmd() not implemented")
}
