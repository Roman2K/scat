package aprocs

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	"scat"
)

type compress struct {
	// TODO level
}

func NewCompress() ProcUnprocer {
	return compress{}
}

func (c compress) Proc() Proc {
	return ChunkFunc(c.process)
}

func (c compress) Unproc() Proc {
	return ChunkFunc(c.unprocess)
}

func (compress) process(c scat.Chunk) (new scat.Chunk, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(c.Data())))
	w := gzip.NewWriter(buf)
	_, err = w.Write(c.Data())
	if err != nil {
		return
	}
	err = w.Close()
	new = c.WithData(buf.Bytes())
	return
}

func (compress) unprocess(c scat.Chunk) (new scat.Chunk, err error) {
	r, err := gzip.NewReader(bytes.NewReader(c.Data()))
	if err != nil {
		return
	}
	buf, err := ioutil.ReadAll(r)
	new = c.WithData(buf)
	return
}
