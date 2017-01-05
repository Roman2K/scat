package tmpdedup

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type Dir struct {
	path    string
	man     *Man
	delOnce sync.Once
}

func TempDir(parent string) (dir *Dir, err error) {
	path, err := ioutil.TempDir(parent, "")
	if err != nil {
		return
	}
	dir = &Dir{
		path: path,
		man:  NewMan(),
	}
	return
}

func (dir *Dir) Get(name string, create func(string) error) (
	path string, wg *sync.WaitGroup, err error,
) {
	path = filepath.Join(dir.path, name)
	manCreate := func() error {
		return create(path)
	}
	wg, err = dir.man.Get(path, manCreate)
	return
}

func (dir *Dir) Finish() (err error) {
	if dir.man.Len() > 0 {
		return errors.New(fmt.Sprintf("leftover files in %s", dir.path))
	}
	dir.delOnce.Do(func() {
		err = os.Remove(dir.path)
	})
	return
}

func (dir *Dir) TmpMan() *Man {
	return dir.man
}
