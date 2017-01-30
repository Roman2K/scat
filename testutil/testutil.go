package testutil

import (
	"scat"
	"scat/checksum"
	"scat/procs"
	"scat/stores"
	"sort"
)

func ReadChunks(ch <-chan procs.Res) (chunks []*scat.Chunk, err error) {
	for res := range ch {
		if e := res.Err; e != nil && err == nil {
			err = e
		}
		chunks = append(chunks, res.Chunk)
	}
	return
}

type FinishErrProc struct {
	Err error
}

var _ procs.Proc = FinishErrProc{}

func (p FinishErrProc) Process(*scat.Chunk) <-chan procs.Res {
	panic("Process() not implemented")
}

func (p FinishErrProc) Finish() error {
	return p.Err
}

func SortCopiersByIdString(s []stores.Copier) (res []stores.Copier) {
	res = make([]stores.Copier, len(s))
	copy(res, s)
	idStr := func(i int) string {
		return res[i].Id().(string)
	}
	sort.Slice(res, func(i, j int) bool {
		return idStr(i) < idStr(j)
	})
	return
}

var Hash1 = struct {
	Hash checksum.Hash
	Hex  string
}{
	Hash: checksum.Hash{
		44, 242, 77, 186, 95, 176, 163, 14, 38, 232, 59, 42, 197, 185, 226, 158,
		27, 22, 30, 92, 31, 167, 66, 94, 115, 4, 51, 98, 147, 139, 152, 36,
	},
	Hex: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
}
