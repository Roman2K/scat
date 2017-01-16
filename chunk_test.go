package scat_test

import (
	"io/ioutil"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/checksum"
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

	c = scat.NewChunk(1, scat.BytesData("yy"))
	c.SetHash(checksum.SumBytes([]byte("some hash")))
	c.Meta().Set("x", "y")
	cHash := c.Hash()

	dup := c.WithData(scat.BytesData("zzz"))
	dup.SetTargetSize(3)
	b, err = dup.Data().Bytes()
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
