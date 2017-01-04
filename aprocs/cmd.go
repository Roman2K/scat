package aprocs

import (
	"bytes"
	"os/exec"

	ss "secsplit"
)

type CmdInFunc func(*ss.Chunk) (*exec.Cmd, error)

var _ Proc = CmdInFunc(nil)

func (fn CmdInFunc) Process(c *ss.Chunk) <-chan Res {
	return InplaceProcFunc(fn.process).Process(c)
}

func (fn CmdInFunc) process(c *ss.Chunk) (err error) {
	cmd, err := fn(c)
	if err != nil {
		return
	}
	cmd.Stdin = bytes.NewReader(c.Data)
	return cmd.Run()
}

func (fn CmdInFunc) Finish() error {
	return nil
}

type CmdOutFunc func(*ss.Chunk) (*exec.Cmd, error)

var _ Proc = CmdOutFunc(nil)

func (fn CmdOutFunc) Process(c *ss.Chunk) <-chan Res {
	return InplaceProcFunc(fn.process).Process(c)
}

func (fn CmdOutFunc) process(c *ss.Chunk) (err error) {
	cmd, err := fn(c)
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

func (fn CmdOutFunc) Finish() error {
	return nil
}
