package scat_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/checksum"
	assert "github.com/stretchr/testify/require"
)

func TestChunk(t *testing.T) {
	c := scat.NewChunk(0, nil)
	assert.Equal(t, 0, c.Num())
	b, err := c.Data().Bytes()
	assert.NoError(t, err)
	assert.Equal(t, "", string(b))
	assert.Equal(t, checksum.Hash{}, c.Hash())
	assert.Equal(t, 0, c.TargetSize())
	assert.Equal(t, nil, c.Meta().Get("xx"))
}

func TestChunkWithData(t *testing.T) {
	c := scat.NewChunk(1, scat.BytesData("yy"))
	c.SetHash(checksum.SumBytes([]byte("some hash")))
	c.Meta().Set("x", "y")
	cHash := c.Hash()

	dup := c.WithData(scat.BytesData("zzz"))
	dup.SetTargetSize(3)
	b, err := dup.Data().Bytes()
	assert.NoError(t, err)
	assert.Equal(t, "zzz", string(b))
	assert.Equal(t, cHash, dup.Hash())
	assert.Equal(t, 0, c.TargetSize())
	assert.Equal(t, 3, dup.TargetSize())
	dup.SetHash(checksum.SumBytes([]byte("some other hash")))
	assert.NotEqual(t, cHash, dup.Hash())
	assert.Equal(t, cHash, c.Hash())
	assert.Equal(t, "y", dup.Meta().Get("x"))
	dup.Meta().Set("x", "y2")
	assert.Equal(t, "y", c.Meta().Get("x"))
}

func TestChunkWithDataNoMeta(t *testing.T) {
	c := scat.NewChunk(9, nil)
	assert.NotPanics(t, func() {
		// ...trying to access nil meta map mutex or something
		c.WithData(scat.BytesData("a"))
	})
}

func TestBytesData(t *testing.T) {
	var bd scat.Data

	bd = scat.BytesData{}
	b, err := bd.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, "", string(b))
	assert.Equal(t, 0, bd.(scat.Sizer).Size())

	bd = scat.BytesData{'a'}
	b, err = bd.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, "a", string(b))
	assert.Equal(t, 1, bd.(scat.Sizer).Size())
}

func TestReaderData(t *testing.T) {
	rd := scat.NewReaderData(strings.NewReader("a"))
	buf, err := ioutil.ReadAll(rd.Reader())
	assert.NoError(t, err)
	assert.Equal(t, "a", string(buf))

	rd = scat.NewReaderData(strings.NewReader("a"))
	rd.Reader()
	assert.Panics(t, func() {
		rd.Reader()
	})

	rd = scat.NewReaderData(strings.NewReader("a"))
	buf, err = rd.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, "a", string(buf))
	assert.Panics(t, func() {
		rd.Bytes()
	})
}
