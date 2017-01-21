package argparse_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat/argparse"
)

func TestArgs(t *testing.T) {
	args := argparse.Args{argparse.ArgStr}

	str := ""
	_, _, err := args.Parse(str)
	assert.Equal(t, argparse.ErrTooFewArgs, err)

	str = " "
	_, _, err = args.Parse(str)
	assert.Equal(t, argparse.ErrTooFewArgs, err)

	str = "abc"
	res, n, err := args.Parse(str)
	assert.NoError(t, err)
	vals := res.([]interface{})
	assert.Equal(t, 1, len(vals))
	assert.Equal(t, "abc", vals[0].(string))
	assert.Equal(t, len(str), n)

	str = " abc "
	res, n, err = args.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 1, len(vals))
	assert.Equal(t, "abc", vals[0].(string))
	assert.Equal(t, len(str), n)

	str = "abc abc"
	_, _, err = args.Parse(str)
	assert.Equal(t, argparse.ErrTooManyArgs, err)
}

func TestArgsWithVariadic(t *testing.T) {
	arg := argparse.Args{argparse.ArgVariadic{argparse.ArgStr}}
	res, n, err := arg.Parse(" ")
	vals := res.([]interface{})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(vals))
	assert.Equal(t, 0, len(vals[0].([]interface{})))
	assert.Equal(t, 1, n)

	arg = argparse.Args{argparse.ArgStr, argparse.ArgVariadic{argparse.ArgStr}}
	res, n, err = arg.Parse("a")
	vals = res.([]interface{})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "a", vals[0])
	assert.Equal(t, 0, len(vals[1].([]interface{})))
	assert.Equal(t, 1, n)
}

func TestArgsEmptyErr(t *testing.T) {
	someErr := errors.New("some err")
	arg := argparse.Args{argEmptyErr{someErr}}
	_, _, err := arg.Parse("")
	assert.Equal(t, someErr, err)
}
