package aprocs_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
)

func TestChain(t *testing.T) {
	a := func(c *ss.Chunk) error {
		c.Data = append(c.Data, 'a')
		return nil
	}
	b := func(c *ss.Chunk) error {
		c.Data = append(c.Data, 'b')
		return nil
	}
	chain := aprocs.NewChain([]aprocs.Proc{
		aprocs.InplaceProcFunc(a),
		aprocs.InplaceProcFunc(b),
	})
	ch := chain.Process(&ss.Chunk{Data: []byte{'x'}})
	res := <-ch
	_, ok := <-ch
	assert.False(t, ok)
	assert.NoError(t, res.Err)
	assert.Equal(t, "xab", string(res.Chunk.Data))
}
