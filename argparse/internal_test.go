package argparse

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestInternalCountLeftSpaces(t *testing.T) {
	for _, test := range [...]struct {
		str   string
		count int
	}{
		{"abc d", 0},
		{" abc d", 1},
		{"  \n abc d", 4},
		{"  \n ", 4},
		{" ", 1},
		{"", 0},
	} {
		assert.Equal(t, test.count, countLeftSpaces(test.str))
	}
}

func TestInternalShorten(t *testing.T) {
	origEtc := etc
	etc = "..."
	defer func() {
		etc = origEtc
	}()
	assert.NotEqual(t, origEtc, etc)

	for i, test := range [...]struct {
		str  string
		pos  int
		str2 string
		pos2 int
	}{
		{"abc", 1, "abc", 1},
		{"a\nbc", 0, "a...", 0},
		{"a\nbc", 1, "a...", 1},
		{"a\nbc\nd", 2, "...bc...", 3},
		{"a\nbc\nd", 3, "...bc...", 4},
		{"\n", 0, "...", 0},
		{"\n", 99, "...", 99 - 1 + 3},
	} {
		str, pos := shorten(test.str, test.pos)
		assert.Equal(t, test.str2, str, "test %d", i)
		assert.Equal(t, test.pos2, pos, "test %d", i)
	}
}
