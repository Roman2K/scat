package cpprocs

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
)

type cat struct {
	dir string
}

func NewCat(dir string) LsProcUnprocer {
	return cat{dir: dir}
}

func (cat cat) Proc() aprocs.Proc {
	return aprocs.CmdInFunc(cat.procCmd)
}

func (cat cat) procCmd(c *ss.Chunk) (cmd *exec.Cmd, err error) {
	path := cat.filePath(c.Hash)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	cmd = exec.Command("cat")
	cmd.Stdout = f
	return
}

func (cat cat) Unproc() aprocs.Proc {
	return aprocs.CmdOutFunc(cat.unprocCmd)
}

func (cat cat) unprocCmd(c *ss.Chunk) (*exec.Cmd, error) {
	path := cat.filePath(c.Hash)
	return exec.Command("cat", path), nil
}

func (cat cat) filePath(hash checksum.Hash) string {
	return filepath.Join(cat.dir, fmt.Sprintf("%x", hash))
}

func (cat cat) Ls() (entries []LsEntry, err error) {
	files, err := ioutil.ReadDir(cat.dir)
	if err != nil {
		return
	}
	var (
		buf   = make([]byte, checksum.Size)
		entry LsEntry
	)
	for _, f := range files {
		n, err := fmt.Sscanf(f.Name(), "%x", &buf)
		if err != nil || n != 1 {
			continue
		}
		err = entry.Hash.LoadSlice(buf)
		if err != nil {
			continue
		}
		fi, err := os.Stat(filepath.Join(cat.dir, f.Name()))
		if err != nil {
			return nil, err
		}
		entry.Size = fi.Size()
		entries = append(entries, entry)
	}
	return
}
