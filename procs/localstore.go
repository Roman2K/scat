package procs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	ss "secsplit"
)

type LocalStore struct {
	Dir string
}

func (s *LocalStore) Proc() Proc {
	return inplaceProcFunc(s.process)
}

func (s *LocalStore) Unproc() Proc {
	return inplaceProcFunc(s.unprocess)
}

func (s *LocalStore) process(c *ss.Chunk) (err error) {
	path := s.path(c)
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(c.Data)
	return
}

func (s *LocalStore) unprocess(c *ss.Chunk) (err error) {
	path := s.path(c)
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	c.Data, err = ioutil.ReadAll(f)
	return
}

func (s *LocalStore) path(c *ss.Chunk) string {
	return filepath.Join(s.Dir, fmt.Sprintf("%x", c.Hash))
}
