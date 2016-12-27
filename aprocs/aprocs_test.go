package aprocs_test

import (
	"fmt"
	"sort"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
)

func testChunkNums(t *testing.T, proc aprocs.Proc, inChunks int) {
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
		for res := range proc.Process(c) {
			assert.NoError(t, res.Err)
			nums = append(nums, res.Chunk.Num)
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

func getErr(t *testing.T, ch <-chan aprocs.Res) error {
	res := <-ch
	_, ok := <-ch
	assert.False(t, ok)
	return res.Err
}

func readChunks(ch <-chan aprocs.Res) (chunks []*ss.Chunk, err error) {
	for res := range ch {
		err = res.Err
		if err != nil {
			return
		}
		chunks = append(chunks, res.Chunk)
	}
	return
}
