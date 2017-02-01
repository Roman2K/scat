package testutil

import (
	"errors"
	"sort"
	"testing"

	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/checksum"
	"github.com/Roman2K/scat/procs"

	"github.com/stretchr/testify/assert"
)

//
// Generate hashes in Ruby with:
//
//		digest  = Digest::SHA256.digest("foo")
//		hex     = digest.unpack("H*")
//		hash    = digest.unpack("C*")
//
var Hashes = [...]struct {
	Hex  string
	Hash checksum.Hash
}{{
	Hex: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
	Hash: checksum.Hash{
		44, 242, 77, 186, 95, 176, 163, 14, 38, 232, 59, 42, 197, 185, 226, 158,
		27, 22, 30, 92, 31, 167, 66, 94, 115, 4, 51, 98, 147, 139, 152, 36,
	},
}, {
	Hex: "cd2eb0837c9b4c962c22d2ff8b5441b7b45805887f051d39bf133b583baf6860",
	Hash: checksum.Hash{
		205, 46, 176, 131, 124, 155, 76, 150, 44, 34, 210, 255, 139, 84, 65, 183,
		180, 88, 5, 136, 127, 5, 29, 57, 191, 19, 59, 88, 59, 175, 104, 96,
	},
}}

var Hash1 = Hashes[0]

func init() {
	hex := func(i int) string {
		return Hashes[i].Hex
	}
	sort.Slice(Hashes[:], func(i, j int) bool {
		return hex(i) < hex(j)
	})
}

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

type Finisher interface {
	Finish() error
}

type getFinisherFn func(procs.Proc) Finisher

func TestFinishErrForward(t *testing.T, getFinisher getFinisherFn) {
	proc := FinishErrProc{Err: nil}
	err := getFinisher(proc).Finish()
	assert.NoError(t, err)
	err = getFinisher(proc).Finish() // idempotence
	assert.NoError(t, err)

	someErr := errors.New("some err")
	proc = FinishErrProc{Err: someErr}
	err = getFinisher(proc).Finish()
	assert.Equal(t, someErr, err)
	err = getFinisher(proc).Finish() // idempotence
	assert.Equal(t, someErr, err)
}
