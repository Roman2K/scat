package argparse_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"
	"github.com/Roman2K/scat/argparse"
)

func TestArgBytes(t *testing.T) {
	str := "1kib"
	i, n, err := argparse.ArgBytes.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1024), i)
	assert.Equal(t, 4, n)

	str = "1kib "
	i, n, err = argparse.ArgBytes.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1024), i)
	assert.Equal(t, 4, n)

	str = " 1kib"
	_, _, err = argparse.ArgBytes.Parse(str)
	assert.Equal(t, argparse.ErrInvalidSyntax, err)
}
