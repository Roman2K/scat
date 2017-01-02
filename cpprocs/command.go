package cpprocs

import (
	"bytes"
	"os/exec"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
)

type CmdSpawner interface {
	NewCmd(checksum.Hash) (*exec.Cmd, error)
}

func NewCommand(spawner CmdSpawner) aprocs.Proc {
	return aprocs.InplaceProcFunc(cmdProc{spawner: spawner}.process)
}

type cmdProc struct {
	spawner CmdSpawner
}

func (cmdp cmdProc) process(c *ss.Chunk) (err error) {
	cmd, err := cmdp.spawner.NewCmd(c.Hash)
	if err != nil {
		return
	}
	cmd.Stdin = bytes.NewReader(c.Data)
	return cmd.Run()
}
