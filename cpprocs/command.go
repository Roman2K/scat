package cpprocs

import (
	"bytes"
	"os/exec"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
)

type cmdProc struct {
	id      interface{}
	spawner CmdSpawner
}

type CmdSpawner interface {
	NewCmd(checksum.Hash) (*exec.Cmd, error)
	Ls() ([]checksum.Hash, error)
}

func NewCommand(id interface{}, spawner CmdSpawner) Proc {
	return cmdProc{id: id, spawner: spawner}
}

func (cmdp cmdProc) Id() interface{} {
	return cmdp.id
}

func (cmdp cmdProc) Ls() ([]checksum.Hash, error) {
	return cmdp.spawner.Ls()
}

func (cmdp cmdProc) Process(c *ss.Chunk) <-chan aprocs.Res {
	return aprocs.InplaceProcFunc(cmdp.process).Process(c)
}

func (cmdp cmdProc) process(c *ss.Chunk) (err error) {
	cmd, err := cmdp.spawner.NewCmd(c.Hash)
	if err != nil {
		return
	}
	cmd.Stdin = bytes.NewReader(c.Data)
	return cmd.Run()
}

func (cmdp cmdProc) Finish() error {
	return nil
}
