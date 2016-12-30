package testutil

import (
	ss "secsplit"
	"secsplit/aprocs"
)

func ReadChunks(ch <-chan aprocs.Res) (chunks []*ss.Chunk, err error) {
	for res := range ch {
		if e := res.Err; e != nil && err == nil {
			err = e
		}
		chunks = append(chunks, res.Chunk)
	}
	return
}
