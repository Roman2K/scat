package stores

import (
	"fmt"
	"path/filepath"

	"github.com/Roman2K/scat/checksum"
)

type Dir struct {
	Path string
	Part StrPart
}

func (d Dir) FullPath(hash checksum.Hash) string {
	filename := fmt.Sprintf("%x", hash)
	parts := append(
		append([]string{d.Path}, d.Part.Split(filename)...),
		filename,
	)
	return filepath.Join(parts...)
}

type DirLister interface {
	Ls(dir string, depth int) <-chan DirLsRes
}

type DirLsRes struct {
	Name string
	Size int64
	Err  error
}

func (d Dir) Ls(lser DirLister) ([]LsEntry, error) {
	ch := lser.Ls(d.Path, len(d.Part)+1)
	collect := func() (entries []LsEntry, err error) {
		var (
			entry LsEntry
			buf   = make([]byte, checksum.Size)
		)
		for res := range ch {
			err = res.Err
			if err != nil {
				return
			}
			n, err := fmt.Sscanf(res.Name, "%x", &buf)
			if err != nil || n != 1 {
				continue
			}
			err = entry.Hash.LoadSlice(buf)
			if err != nil {
				continue
			}
			entry.Size = res.Size
			entries = append(entries, entry)
		}
		return
	}
	entries, err := collect()
	for range ch {
	}
	return entries, err
}
