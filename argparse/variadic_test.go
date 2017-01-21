package argparse_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat/argparse"
)

func TestArgVariadic(t *testing.T) {
	arg := argparse.ArgVariadic{argparse.ArgStr}

	str := ""
	res, n, err := arg.Parse(str)
	assert.NoError(t, err)
	vals := res.([]interface{})
	assert.Equal(t, 0, len(vals))
	assert.Equal(t, len(str), n)

	str = "x"
	res, n, err = arg.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 1, len(vals))
	assert.Equal(t, len(str), n)
	assert.Equal(t, "x", vals[0].(string))

	str = "x y"
	res, n, err = arg.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, len(str), n)
	assert.Equal(t, "x", vals[0].(string))
	assert.Equal(t, "y", vals[1].(string))

	str = " x y "
	res, n, err = arg.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, len(str), n)
	assert.Equal(t, "x", vals[0].(string))
	assert.Equal(t, "y", vals[1].(string))

	// arg err
	someErr := errors.New("some err")
	arg = argparse.ArgVariadic{argErr{someErr}}
	_, _, err = arg.Parse("x")
	assert.Equal(t, someErr, err)
}
