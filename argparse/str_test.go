package argparse_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/Roman2K/scat/argparse"
)

func TestArgStr(t *testing.T) {
	res, n, err := argparse.ArgStr.Parse("abc")
	assert.NoError(t, err)
	assert.Equal(t, "abc", res.(string))
	assert.Equal(t, 3, n)

	res, n, err = argparse.ArgStr.Parse(" abc ")
	assert.NoError(t, err)
	assert.Equal(t, "", res.(string))
	assert.Equal(t, 0, n)

	res, n, err = argparse.ArgStr.Parse("  ")
	assert.NoError(t, err)
	assert.Equal(t, "", res.(string))
	assert.Equal(t, 0, n)
}
