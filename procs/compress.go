package procs

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	ss "secsplit"
)

type Compress struct {
	// TODO level
}

func (c *Compress) Proc() Proc {
	return inplaceProcFunc(c.process)
}

func (c *Compress) Unproc() Proc {
	return inplaceProcFunc(c.unprocess)
}

func (*Compress) process(chunk *ss.Chunk) (err error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(chunk.Data)))
	w := gzip.NewWriter(buf)
	_, err = w.Write(chunk.Data)
	if err != nil {
		return
	}
	err = w.Close()
	if err != nil {
		return
	}
	chunk.Data = buf.Bytes()
	return
}

func (*Compress) unprocess(chunk *ss.Chunk) (err error) {
	r, err := gzip.NewReader(bytes.NewReader(chunk.Data))
	if err != nil {
		return
	}
	chunk.Data, err = ioutil.ReadAll(r)
	return
}
