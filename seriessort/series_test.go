package seriessort_test

import (
	"testing"

	"github.com/Roman2K/scat/seriessort"
	assert "github.com/stretchr/testify/require"
)

func TestSeries(t *testing.T) {
	s := seriessort.Series{}
	sorted := s.Sorted()
	assert.Equal(t, 0, len(sorted))
	assert.Equal(t, 0, s.Len())

	s.Add(1, "b")
	sorted = s.Sorted()
	assert.Equal(t, 0, len(sorted))
	assert.Equal(t, 2, s.Len())

	s.Add(2, "c")
	sorted = s.Sorted()
	assert.Equal(t, 0, len(sorted))
	assert.Equal(t, 3, s.Len())

	s.Add(0, "a")
	sorted = s.Sorted()
	assert.Equal(t, 3, len(sorted))
	assert.Equal(t, 3, s.Len())
	assert.Equal(t, "a", sorted[0].(string))
	assert.Equal(t, "b", sorted[1].(string))
	assert.Equal(t, "c", sorted[2].(string))

	s.Drop(0)
	sorted = s.Sorted()
	assert.Equal(t, 3, len(sorted))
	assert.Equal(t, 3, s.Len())
	assert.Equal(t, "a", sorted[0].(string))

	s.Drop(1)
	sorted = s.Sorted()
	assert.Equal(t, 2, len(sorted))
	assert.Equal(t, 2, s.Len())
	assert.Equal(t, "b", sorted[0].(string))

	s.Drop(1)
	sorted = s.Sorted()
	assert.Equal(t, 1, len(sorted))
	assert.Equal(t, 1, s.Len())
	assert.Equal(t, "c", sorted[0].(string))

	s.Add(3, "d")
	sorted = s.Sorted()
	assert.Equal(t, 2, len(sorted))
	assert.Equal(t, 2, s.Len())
	assert.Equal(t, "c", sorted[0].(string))
	assert.Equal(t, "d", sorted[1].(string))

	s.Add(2, "x")
	sorted = s.Sorted()
	assert.Equal(t, 2, len(sorted))
	assert.Equal(t, 2, s.Len())
	assert.Equal(t, "x", sorted[0].(string))
	assert.Equal(t, "d", sorted[1].(string))

	assert.Panics(t, func() {
		s.Add(0, "x")
	})
	assert.Panics(t, func() {
		s.Add(1, "x")
	})
	sorted = s.Sorted()
	assert.Equal(t, 2, len(sorted))
	assert.Equal(t, 2, s.Len())

	s.Drop(-1)
	sorted = s.Sorted()
	assert.Equal(t, 2, len(sorted))
	assert.Equal(t, 2, s.Len())
	s.Drop(99)
	sorted = s.Sorted()
	assert.Equal(t, 0, len(sorted))
	assert.Equal(t, 0, s.Len())
	s.Drop(99)
	sorted = s.Sorted()
	assert.Equal(t, 0, len(sorted))
	assert.Equal(t, 0, s.Len())
}
