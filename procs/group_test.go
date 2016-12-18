package procs

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
)

func TestGroup(t *testing.T) {
	g := Group(2)

	res := g.Process(&ss.Chunk{Num: 1})
	assert.NoError(t, res.Err)
	assert.Equal(t, 0, len(res.Chunks))

	res = g.Process(&ss.Chunk{Num: 2})
	assert.NoError(t, res.Err)
	assert.Equal(t, 0, len(res.Chunks))

	res = g.Process(&ss.Chunk{Num: 0})
	assert.NoError(t, res.Err)
	assert.Equal(t, 1, len(res.Chunks))
	assert.Equal(t, 1, len(g.(*group).growing))

	chunk := res.Chunks[0]
	assert.Equal(t, 1, chunk.Num)
	chunks, ok := chunk.GetMeta("group").([]*ss.Chunk)
	assert.True(t, ok)
	assert.Equal(t, 2, len(chunks))
	assert.Equal(t, 0, chunks[0].Num)
	assert.Equal(t, 1, chunks[1].Num)

	res = g.Process(&ss.Chunk{Num: 3})
	assert.NoError(t, res.Err)
	assert.Equal(t, 1, len(res.Chunks))
	assert.Equal(t, 0, len(g.(*group).growing))

	chunk = res.Chunks[0]
	assert.Equal(t, 3, chunk.Num)
	chunks, ok = chunk.GetMeta("group").([]*ss.Chunk)
	assert.True(t, ok)
	assert.Equal(t, 2, len(chunks))
	assert.Equal(t, 2, chunks[0].Num)
	assert.Equal(t, 3, chunks[1].Num)
}

func TestGroupMinSize(t *testing.T) {
	assert.Panics(t, func() { Group(0) })
	assert.NotPanics(t, func() { Group(1) })
}
