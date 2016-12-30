package copies_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
	"secsplit/cpprocs"
	"secsplit/cpprocs/copies"
)

func TestCopies(t *testing.T) {
	reg := copies.NewReg()
	hash1 := checksum.Sum([]byte("chunk1"))
	a := testCpProc{
		id:     "a",
		hashes: []checksum.Hash{hash1},
	}
	b := testCpProc{
		id:     "b",
		hashes: []checksum.Hash{hash1},
	}
	err := reg.Add([]cpprocs.Proc{a})
	assert.NoError(t, err)
	assert.Equal(t, 1, reg.List(hash1).UnlockedLen())
	assert.True(t, reg.List(hash1).UnlockedContains(a))
	assert.False(t, reg.List(hash1).UnlockedContains(b))
}

type testCpProc struct {
	id     interface{}
	hashes []checksum.Hash
}

func (cpp testCpProc) Id() interface{} {
	return cpp.id
}

func (cpp testCpProc) Ls() ([]checksum.Hash, error) {
	return cpp.hashes, nil
}

func (cpp testCpProc) Process(c *ss.Chunk) <-chan aprocs.Res {
	return nil
}

func (cpp testCpProc) Finish() error {
	return nil
}
