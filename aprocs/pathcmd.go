package aprocs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	ss "secsplit"
	"secsplit/tmpdedup"
)

type pathCmdIn struct {
	newCmd     newPathCmdIn
	tmp        string
	tmpMan     *tmpdedup.Man
	tmpDeleted bool
}

type newPathCmdIn func(*ss.Chunk, string) (*exec.Cmd, error)

func NewPathCmdIn(tmpDir string, newCmd newPathCmdIn) (proc Proc, err error) {
	tmp, err := ioutil.TempDir(tmpDir, "")
	proc = &pathCmdIn{
		newCmd: newCmd,
		tmp:    tmp,
		tmpMan: tmpdedup.NewMan(),
	}
	return
}

func (cmdp *pathCmdIn) TmpMan() *tmpdedup.Man {
	return cmdp.tmpMan
}

func (cmdp *pathCmdIn) Process(c *ss.Chunk) <-chan Res {
	return InplaceProcFunc(cmdp.process).Process(c)
}

func (cmdp *pathCmdIn) process(c *ss.Chunk) (err error) {
	path := filepath.Join(cmdp.tmp, fmt.Sprintf("%x", c.Hash))
	wg, err := cmdp.tmpMan.Get(path, func() (err error) {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return
		}
		defer f.Close()
		_, err = f.Write(c.Data)
		return
	})
	if err != nil {
		return
	}
	defer wg.Done()
	cmd, err := cmdp.newCmd(c, path)
	if err != nil {
		return
	}
	return cmd.Run()
}

func (cmdp *pathCmdIn) Finish() (err error) {
	if cmdp.tmpMan.Len() > 0 {
		return errors.New(fmt.Sprintf("leftover files in %s", cmdp.tmp))
	}
	if !cmdp.tmpDeleted {
		err = os.Remove(cmdp.tmp)
		cmdp.tmpDeleted = true
	}
	return
}
