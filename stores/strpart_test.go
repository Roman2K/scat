package stores_test

import (
	"testing"

	"github.com/Roman2K/scat/stores"
	assert "github.com/stretchr/testify/require"
)

func TestStrPart(t *testing.T) {
	part := stores.StrPart{}
	parts := part.Split("abc")
	assert.Equal(t, []string{}, parts)

	part = stores.StrPart{0, 0}
	parts = part.Split("abc")
	assert.Equal(t, []string{"", ""}, parts)

	part = stores.StrPart{2}
	parts = part.Split("abc")
	assert.Equal(t, []string{"ab"}, parts)

	part = stores.StrPart{4}
	parts = part.Split("abc")
	assert.Equal(t, []string{"abc"}, parts)

	part = stores.StrPart{1, 1}
	parts = part.Split("abc")
	assert.Equal(t, []string{"a", "b"}, parts)

	part = stores.StrPart{2, 2}
	parts = part.Split("abc")
	assert.Equal(t, []string{"ab", "c"}, parts)
}
