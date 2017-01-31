package stores

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"scat"
	"scat/procs"
)

type Cp struct {
	Dir Dir
}

var _ Store = Cp{}

func (cp Cp) Proc() procs.Proc {
	return procs.InplaceFunc(cp.process)
}

func (cp Cp) process(c *scat.Chunk) (err error) {
	path := cp.Dir.FullPath(c.Hash())
	open := func() (io.WriteCloser, error) {
		return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	}
	w, err := open()
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return
		}
		w, err = open()
	}
	if err != nil {
		return
	}
	defer w.Close()
	_, err = io.Copy(w, c.Data().Reader())
	return
}

func (cp Cp) Unproc() procs.Proc {
	return procs.ChunkFunc(cp.unprocess)
}

func (cp Cp) unprocess(c *scat.Chunk) (new *scat.Chunk, err error) {
	path := cp.Dir.FullPath(c.Hash())
	b, err := ioutil.ReadFile(path)
	new = c.WithData(scat.BytesData(b))
	return
}

func (cp Cp) Ls() ([]LsEntry, error) {
	return cp.Dir.Ls(localLister{})
}

type localLister struct{}

func (localLister) Ls(dir string, depth int) <-chan DirLsRes {
	_, err := os.Stat(dir)
	if err != nil {
		ch := make(chan DirLsRes, 1)
		defer close(ch)
		ch <- DirLsRes{Err: err}
		return ch
	}
	pattern := depthPattern(dir, depth)
	ch := make(chan DirLsRes)
	go func() {
		defer close(ch)
		sendResults := func() error {
			files, err := filepath.Glob(pattern)
			if err != nil {
				return err
			}
			for _, path := range files {
				fi, err := os.Stat(path)
				if err != nil {
					return err
				}
				if !fi.IsDir() {
					ch <- DirLsRes{Name: fi.Name(), Size: fi.Size()}
				}
			}
			return nil
		}
		err := sendResults()
		if err != nil {
			ch <- DirLsRes{Err: err}
		}
	}()
	return ch
}

func depthPattern(path string, depth int) string {
	wildcards := make([]string, depth)
	for i := range wildcards {
		wildcards[i] = "*"
	}
	return filepath.Join(append([]string{path}, wildcards...)...)
}
