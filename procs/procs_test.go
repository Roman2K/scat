package procs_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/checksum"
	"github.com/Roman2K/scat/procs"
	assert "github.com/stretchr/testify/require"
)

func testChunkNums(t *testing.T, proc procs.Proc, inChunks int) {
	newChunk := func(num int) (c *scat.Chunk) {
		data := scat.BytesData{'a'}
		c = scat.NewChunk(num, data)
		c.SetHash(checksum.SumBytes(data))
		return
	}

	nums := []int{}
	for i := inChunks - 1; i >= 0; i-- {
		c := newChunk(i)
		for res := range proc.Process(c) {
			assert.NoError(t, res.Err)
			nums = append(nums, res.Chunk.Num())
		}
	}
	sort.Ints(nums)
	assert.True(t, contiguousInts(nums), fmt.Sprintf("not contiguous: %v", nums))
	assert.Equal(t, 0, nums[0])
}

func contiguousInts(ints []int) bool {
	for i, n := 1, len(ints); i < n; i++ {
		if ints[i-1] != ints[i]-1 {
			return false
		}
	}
	return true
}

func getErr(t *testing.T, ch <-chan procs.Res) error {
	res := <-ch
	_, ok := <-ch
	assert.False(t, ok)
	return res.Err
}
