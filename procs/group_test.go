package procs_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/procs"
	"scat/testutil"
)

func TestGroup(t *testing.T) {
	g := procs.NewGroup(2)

	chunks, err := testutil.ReadChunks(g.Process(scat.NewChunk(1, nil)))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(chunks))

	chunks, err = testutil.ReadChunks(g.Process(scat.NewChunk(2, nil)))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(chunks))

	chunks, err = testutil.ReadChunks(g.Process(scat.NewChunk(0, nil)))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))

	chunk := chunks[0]
	assert.Equal(t, 0, chunk.Num())
	grp := chunk.Meta().Get("group").([]*scat.Chunk)
	assert.Equal(t, 2, len(grp))
	assert.Equal(t, 0, grp[0].Num())
	assert.Equal(t, 1, grp[1].Num())

	chunks, err = testutil.ReadChunks(g.Process(scat.NewChunk(3, nil)))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))

	chunk = chunks[0]
	assert.Equal(t, 1, chunk.Num())
	grp = chunk.Meta().Get("group").([]*scat.Chunk)
	assert.Equal(t, 2, len(grp))
	assert.Equal(t, 2, grp[0].Num())
	assert.Equal(t, 3, grp[1].Num())
}

func TestGroupMinSize(t *testing.T) {
	assert.Panics(t, func() { procs.NewGroup(0) })
	assert.NotPanics(t, func() { procs.NewGroup(1) })
}

func TestGroupChunkNums(t *testing.T) {
	testChunkNums(t, procs.NewGroup(2), 6)
}

func TestGroupFinish(t *testing.T) {
	g := procs.NewGroup(2)

	// 0 ok
	// 1 missing
	_, err := testutil.ReadChunks(g.Process(scat.NewChunk(0, nil)))
	assert.NoError(t, err)
	err = g.Finish()
	assert.Equal(t, procs.ErrShort, err)

	// idempotence
	err = g.Finish()
	assert.Equal(t, procs.ErrShort, err)

	// 0 ok
	// 1 ok
	_, err = testutil.ReadChunks(g.Process(scat.NewChunk(1, nil)))
	assert.NoError(t, err)
	err = g.Finish()
	assert.NoError(t, err)

	// idempotence
	err = g.Finish()
	assert.NoError(t, err)
}
