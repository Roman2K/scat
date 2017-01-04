package copies_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"secsplit/checksum"
	"secsplit/cpprocs/copies"
)

func TestCopies(t *testing.T) {
	reg := copies.NewReg()
	hash1 := checksum.Sum([]byte("hash1"))
	hash2 := checksum.Sum([]byte("hash2"))
	a := owner("a")
	b := owner("b")
	reg.List(hash1).Add(a)
	assert.Equal(t, 1, reg.List(hash1).Len())
	assert.True(t, reg.List(hash1).Contains(a))
	assert.False(t, reg.List(hash2).Contains(a))
	assert.False(t, reg.List(hash1).Contains(b))
	reg.RemoveOwner(a)
	assert.False(t, reg.List(hash1).Contains(a))
	reg.RemoveOwner(b)
	assert.False(t, reg.List(hash1).Contains(b))
}

type owner string

func (o owner) Id() interface{} {
	return o
}
