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
