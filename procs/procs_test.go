package procs

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	ss "secsplit"
	"secsplit/checksum"
)

func testChunkNums(t *testing.T, proc Proc, inChunks int) {
	newChunk := func(num int) *ss.Chunk {
		data := []byte{'a'}
		return &ss.Chunk{
			Num:  num,
			Data: data,
			Hash: checksum.Sum(data),
			Size: len(data),
		}
	}

	nums := []int{}
	for i := inChunks - 1; i >= 0; i-- {
		c := newChunk(i)
		res := proc.Process(c)
		assert.NoError(t, res.Err)
		for _, c := range res.Chunks {
			nums = append(nums, c.Num)
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
