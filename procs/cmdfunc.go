package procs

import (
	"bytes"
	"os/exec"

	"gitlab.com/Roman2K/scat"
)

var (
	_ Proc = CmdFunc(nil)
	_ Proc = CmdInFunc(nil)
	_ Proc = CmdOutFunc(nil)
)

type CmdFunc func(*scat.Chunk) (*exec.Cmd, error)

func (fn CmdFunc) Process(c *scat.Chunk) <-chan Res {
	outFn := CmdOutFunc(func(*scat.Chunk) (cmd *exec.Cmd, err error) {
		cmd, err = fn(c)
		if err != nil {
			return
		}
		cmd.Stdin = c.Data().Reader()
		return
	})
	return outFn.Process(c)
}

func (CmdFunc) Finish() error {
	return nil
}

type CmdInFunc CmdFunc

func (fn CmdInFunc) Process(c *scat.Chunk) <-chan Res {
	return InplaceFunc(fn.process).Process(c)
}

func (fn CmdInFunc) process(c *scat.Chunk) (err error) {
	cmd, err := fn(c)
	if err != nil {
		return
	}
	errBuf := &bytes.Buffer{}
	cmd.Stderr = errBuf
	cmd.Stdin = c.Data().Reader()
	err = cmd.Run()
	if exit, ok := err.(*exec.ExitError); ok {
		exit.Stderr = errBuf.Bytes()
	}
	return
}

func (CmdInFunc) Finish() error {
	return nil
}

type CmdOutFunc CmdFunc

func (fn CmdOutFunc) Process(c *scat.Chunk) <-chan Res {
	return ChunkFunc(fn.process).Process(c)
}

func (fn CmdOutFunc) process(c *scat.Chunk) (new *scat.Chunk, err error) {
	cmd, err := fn(c)
	if err != nil {
		return
	}
	out, err := cmd.Output()
	new = c.WithData(scat.BytesData(out))
	return
}

func (CmdOutFunc) Finish() error {
	return nil
}
