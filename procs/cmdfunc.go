package procs

import (
	"bytes"
	"os/exec"

	"github.com/Roman2K/scat"
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
	cmd.Stdin = c.Data().Reader()
	return runCaptureStderr(cmd)
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
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	err = runCaptureStderr(cmd)
	new = c.WithData(scat.BytesData(buf.Bytes()))
	return
}

func (CmdOutFunc) Finish() error {
	return nil
}

func runCaptureStderr(cmd *exec.Cmd) (err error) {
	var errBuf *bytes.Buffer
	if cmd.Stderr == nil {
		errBuf = &bytes.Buffer{}
		cmd.Stderr = errBuf
	}
	err = cmd.Run()
	if exit, ok := err.(*exec.ExitError); ok {
		if len(exit.Stderr) == 0 && errBuf != nil {
			exit.Stderr = errBuf.Bytes()
		}
	}
	return
}
