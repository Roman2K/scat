package cpprocs

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"secsplit/checksum"
)

type cat struct {
	dir string
}

func NewCat(dir string) CmdSpawner {
	return cat{dir: dir}
}

func (cat cat) NewCmd(hash checksum.Hash) (cmd *exec.Cmd, err error) {
	path := filepath.Join(cat.dir, fmt.Sprintf("%x", hash))
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	cmd = exec.Command("cat")
	cmd.Stdout = f
	return
}

func (cat cat) Ls() (hashes []checksum.Hash, err error) {
	files, err := ioutil.ReadDir(cat.dir)
	if err != nil {
		return
	}
	var (
		buf  []byte
		hash checksum.Hash
	)
	for _, f := range files {
		n, err := fmt.Sscanf(f.Name(), "%x", &buf)
		if err != nil || n != 1 {
			continue
		}
		err = hash.LoadSlice(buf)
		if err != nil {
			continue
		}
		hashes = append(hashes, hash)
	}
	return
}
