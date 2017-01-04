package cpprocs

import (
	"bytes"
	"os/exec"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
)

type command struct {
	spawner CmdSpawner
}

type CmdSpawner interface {
	Lister
	NewProcCmd(checksum.Hash) (*exec.Cmd, error)
	NewUnprocCmd(checksum.Hash) (*exec.Cmd, error)
}

func NewCommand(spawner CmdSpawner) LsProcUnprocer {
	return command{spawner: spawner}
}

func (cmdp command) LsProc() LsProc {
	return NewLsProc(cmdp.spawner, aprocs.InplaceProcFunc(cmdp.process))
}

func (cmdp command) process(c *ss.Chunk) (err error) {
	cmd, err := cmdp.spawner.NewProcCmd(c.Hash)
	if err != nil {
		return
	}
	cmd.Stdin = bytes.NewReader(c.Data)
	return cmd.Run()
}

func (cmdp command) LsUnproc() LsProc {
	return NewLsProc(cmdp.spawner, aprocs.InplaceProcFunc(cmdp.unprocess))
}

func (cmdp command) unprocess(c *ss.Chunk) (err error) {
	cmd, err := cmdp.spawner.NewUnprocCmd(c.Hash)
	if err != nil {
		return
	}
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	err = cmd.Run()
	if err == nil {
		c.Data = buf.Bytes()
	}
	return
}
