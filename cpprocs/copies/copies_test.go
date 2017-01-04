package copies_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"secsplit/checksum"
	"secsplit/cpprocs"
	"secsplit/cpprocs/copies"
	"secsplit/testutil"
)

func TestCopies(t *testing.T) {
	reg := copies.NewReg()
	hash1 := checksum.Sum([]byte("hash1"))
	hash2 := checksum.Sum([]byte("hash2"))
	a := cpprocs.NewCopier("a",
		cpprocs.NewLsProc(testutil.SliceLister{{Hash: hash1}}, nil),
	)
	b := cpprocs.NewCopier("b",
		cpprocs.NewLsProc(testutil.SliceLister{{Hash: hash1}}, nil),
	)
	ls, err := a.Ls()
	assert.NoError(t, err)
	reg.AddCopier(a, ls)
	assert.Equal(t, 1, reg.List(hash1).Len())
	assert.True(t, reg.List(hash1).Contains(a))
	assert.False(t, reg.List(hash2).Contains(a))
	assert.False(t, reg.List(hash1).Contains(b))
}
