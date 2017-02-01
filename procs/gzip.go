package procs

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"

	"scat"
)

type Gzip struct {
	// TODO level
}

func (gz Gzip) Proc() Proc {
	return ChunkFunc(gz.process)
}

func (gz Gzip) Unproc() Proc {
	return ChunkFunc(gz.unprocess)
}

func (Gzip) process(c *scat.Chunk) (new *scat.Chunk, err error) {
	buf := &bytes.Buffer{}
	w := gzip.NewWriter(buf)
	_, err = io.Copy(w, c.Data().Reader())
	if err != nil {
		return
	}
	err = w.Close()
	new = c.WithData(scat.BytesData(buf.Bytes()))
	return
}

func (Gzip) unprocess(c *scat.Chunk) (new *scat.Chunk, err error) {
	r, err := gzip.NewReader(c.Data().Reader())
	if err != nil {
		return
	}
	buf, err := ioutil.ReadAll(r)
	new = c.WithData(scat.BytesData(buf))
	return
}
