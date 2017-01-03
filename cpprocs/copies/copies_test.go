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
	a := cpprocs.Copier{
		Id:     "a",
		Lister: testutil.SliceLister{hash1},
	}
	b := cpprocs.Copier{
		Id:     "b",
		Lister: testutil.SliceLister{hash1},
	}
	err := reg.Add([]cpprocs.Copier{a})
	assert.NoError(t, err)
	assert.Equal(t, 1, reg.List(hash1).Len())
	assert.True(t, reg.List(hash1).Contains(a))
	assert.False(t, reg.List(hash2).Contains(a))
	assert.False(t, reg.List(hash1).Contains(b))
}
