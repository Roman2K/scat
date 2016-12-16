package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplit(t *testing.T) {
	const (
		minSize = 512 * 1024
		maxSize = minSize * 2
	)

	r := &bytes.Buffer{}
	writeN := func(b byte, count int) {
		for i := 0; i < count; i++ {
			r.Write([]byte{b})
		}
	}
	writeN('o', maxSize)
	writeN('k', 1)
	writeN('x', 1)

	split := newSplitterSize(r, minSize, maxSize)
	store := memStore{}
	ppool := newProcPool(1, procChain{
		inplace(store.Process).Process,
	}.Process)
	err := process(split, ppool.Process)
	if err != nil {
		t.Fatal(err)
	}
	err = ppool.Finish()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(store))
	assert.Equal(t, maxSize, len(store[0]))
	assert.Equal(t, 2, len(store[1]))

	all := func(b byte, data []byte) bool {
		for _, a := range data {
			if a != b {
				return false
			}
		}
		return true
	}

	msg := func(data []byte) string {
		extract := 3
		if max := len(data); extract > max {
			extract = max
		}
		start := data[:extract]
		end := data[len(data)-extract:]
		return fmt.Sprintf("len=%d start=%q end=%q", len(data), start, end)
	}

	assert.True(t, all('o', store[0]), msg(store[0]))
	assert.Equal(t, []byte{'k', 'x'}, store[1], msg(store[1]))
}

type memStore map[int][]byte

func (s memStore) Process(c *Chunk) error {
	s[c.Num] = c.Data
	return nil
}
