package stores

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"scat"
	"scat/checksum"
	"scat/procs"
)

type Cp struct {
	Dir  string
	Part StrPart
}

var _ Store = Cp{}

func (cp Cp) Proc() procs.Proc {
	return procs.InplaceFunc(cp.process)
}

func (cp Cp) process(c *scat.Chunk) (err error) {
	path := cp.path(c)
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

func (cp Cp) path(c *scat.Chunk) string {
	filename := fmt.Sprintf("%x", c.Hash())
	parts := append(
		append([]string{cp.Dir}, cp.Part.Split(filename)...),
		filename,
	)
	return filepath.Join(parts...)
}

func (cp Cp) Unproc() procs.Proc {
	return procs.ChunkFunc(cp.unprocess)
}

func (cp Cp) unprocess(c *scat.Chunk) (new *scat.Chunk, err error) {
	b, err := ioutil.ReadFile(cp.path(c))
	new = c.WithData(scat.BytesData(b))
	return
}

func (cp Cp) Ls() (entries []LsEntry, err error) {
	parts := make([]string, len(cp.Part))
	for i, n := 0, len(parts); i < n; i++ {
		parts[i] = "*"
	}
	pattern := filepath.Join(append([]string{cp.Dir}, parts...)...)
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	var (
		buf   = make([]byte, checksum.Size)
		entry LsEntry
	)
	for _, dir := range dirs {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			continue
		}
		old := entries
		entries = make([]LsEntry, len(old), len(old)+len(files))
		copy(entries, old)
		for _, fi := range files {
			n, err := fmt.Sscanf(fi.Name(), "%x", &buf)
			if err != nil || n != 1 {
				continue
			}
			err = entry.Hash.LoadSlice(buf)
			if err != nil {
				continue
			}
			entry.Size = fi.Size()
			entries = append(entries, entry)
		}
	}
	return
}
