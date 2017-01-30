package stores_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat/stores"
)

func TestStrPart(t *testing.T) {
	nest := stores.StrPart{}
	parts := nest.Split("abc")
	assert.Equal(t, []string{}, parts)

	nest = stores.StrPart{0, 0}
	parts = nest.Split("abc")
	assert.Equal(t, []string{"", ""}, parts)

	nest = stores.StrPart{2}
	parts = nest.Split("abc")
	assert.Equal(t, []string{"ab"}, parts)

	nest = stores.StrPart{4}
	parts = nest.Split("abc")
	assert.Equal(t, []string{"abc"}, parts)

	nest = stores.StrPart{1, 1}
	parts = nest.Split("abc")
	assert.Equal(t, []string{"a", "b"}, parts)

	nest = stores.StrPart{2, 2}
	parts = nest.Split("abc")
	assert.Equal(t, []string{"ab", "c"}, parts)
}
