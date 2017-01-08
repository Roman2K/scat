package scat_test

import (
	"scat"
	"scat/checksum"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestChunk(t *testing.T) {
	c := scat.NewChunk(0, nil)
	assert.Equal(t, 0, c.Num())
	assert.Nil(t, c.Data())
	assert.Equal(t, checksum.Hash{}, c.Hash())
	assert.Equal(t, 0, c.TargetSize())
	assert.Equal(t, nil, c.Meta().Get("xx"))

	c = scat.NewChunk(1, []byte("yy"))
	c.SetHash(checksum.Sum([]byte("some hash")))
	c.Meta().Set("x", "y")
	cHash := c.Hash()

	dup := c.WithData([]byte("zzz"))
	assert.Equal(t, []byte("zzz"), dup.Data())
	assert.Equal(t, cHash, dup.Hash())
	assert.Equal(t, 2, c.TargetSize())
	assert.Equal(t, 2, dup.TargetSize())
	dup.SetHash(checksum.Sum([]byte("some other hash")))
	assert.NotEqual(t, cHash, dup.Hash())
	assert.Equal(t, cHash, c.Hash())
	assert.Equal(t, "y", dup.Meta().Get("x"))
	dup.Meta().Set("x", "y2")
	assert.Equal(t, "y", c.Meta().Get("x"))
}
