package argparse_test

import (
	"strconv"
	"testing"

	"gitlab.com/Roman2K/scat/argparse"
	assert "github.com/stretchr/testify/require"
)

func TestArgInt(t *testing.T) {
	str := "1"
	i, n, err := argparse.ArgInt.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, 1, i)
	assert.Equal(t, 1, n)

	str = "1 "
	i, n, err = argparse.ArgInt.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, 1, i)
	assert.Equal(t, 1, n)

	str = " 1"
	_, _, err = argparse.ArgInt.Parse(str)
	assert.Error(t, err)
	assert.IsType(t, &strconv.NumError{}, err)
}
