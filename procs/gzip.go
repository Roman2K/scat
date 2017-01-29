package procs

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"

	"scat"
)

type gzipProc struct {
	// TODO level
}

func NewGzip() ProcUnprocer {
	return gzipProc{}
}

func (gp gzipProc) Proc() Proc {
	return ChunkFunc(gp.process)
}

func (gp gzipProc) Unproc() Proc {
	return ChunkFunc(gp.unprocess)
}

func (gzipProc) process(c *scat.Chunk) (new *scat.Chunk, err error) {
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

func (gzipProc) unprocess(c *scat.Chunk) (new *scat.Chunk, err error) {
	r, err := gzip.NewReader(c.Data().Reader())
	if err != nil {
		return
	}
	buf, err := ioutil.ReadAll(r)
	new = c.WithData(scat.BytesData(buf))
	return
}
