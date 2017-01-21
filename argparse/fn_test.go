package argparse_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat/argparse"
)

func TestArgFn(t *testing.T) {
	received := []([]interface{}){}
	reset := func() {
		received = received[:0]
	}
	someErr := errors.New("some err")

	argFn := argparse.ArgFn{
		"abc": argparse.Fn{
			Run: func(iargs []interface{}) (interface{}, error) {
				received = append(received, iargs)
				return "some str", nil
			},
		},
		"abcerr": argparse.Fn{
			Run: func([]interface{}) (interface{}, error) {
				return nil, someErr
			},
		},
		"xyz": argparse.Fn{
			Args: argparse.Args{argparse.ArgBytes, argparse.ArgBytes},
			Run: func(iargs []interface{}) (interface{}, error) {
				return iargs, nil
			},
		},
	}

	// without args
	reset()
	str := "abc[]"
	res, n, err := argFn.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(received))
	assert.Equal(t, 0, len(received[0]))
	assert.Equal(t, "some str", res.(string))
	assert.Equal(t, len(str), n)

	// fn error
	str = "abcerr"
	_, _, err = argFn.Parse(str)
	assert.Equal(t, someErr, err)

	// too many args
	str = "abc[xxx]"
	_, _, err = argFn.Parse(str)
	assert.Equal(t, argparse.ErrTooManyArgs, err)

	// optional []
	str = "abc"
	res, n, err = argFn.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, "some str", res.(string))
	assert.Equal(t, len(str), n)

	// with args
	str = "xyz[1kib 2kib]"
	res, n, err = argFn.Parse(str)
	assert.NoError(t, err)
	vals := res.([]interface{})
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, uint64(1024), vals[0].(uint64))
	assert.Equal(t, uint64(2048), vals[1].(uint64))
	assert.Equal(t, len(str), n)

	// spaces
	str = "xyz[ 1kib 2kib ]"
	res, n, err = argFn.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, len(str), n)

	// too few args
	str = "xyz[1kib]"
	_, _, err = argFn.Parse(str)
	assert.Equal(t, argparse.ErrTooFewArgs, err)

	// inexistent function
	str = "xxx[]"
	_, _, err = argFn.Parse(str)
	assert.Equal(t, `no such function: "xxx"`, err.Error())
}

func TestArgFnArgErr(t *testing.T) {
	someErr := errors.New("some err")
	argFn := argparse.ArgFn{
		"a": argparse.Fn{
			Args: argparse.Args{argErr{someErr}},
			Run: func([]interface{}) (interface{}, error) {
				return nil, nil
			},
		},
	}
	_, _, err := argFn.Parse("a[b]")
	assert.Equal(t, someErr, err)
}

func TestArgFnNested(t *testing.T) {
	argFn := argparse.ArgFn{}
	argFn["a"] = argparse.Fn{
		Args: argparse.Args{argparse.ArgVariadic{argFn}},
		Run: func(args []interface{}) (interface{}, error) {
			varArgs := args[0].([]interface{})
			return varArgs, nil
		},
	}
	argFn["b"] = argparse.Fn{
		Args: argparse.Args{argparse.ArgStr},
		Run: func(args []interface{}) (interface{}, error) {
			return args[0], nil
		},
	}

	str := "a[b[xxx] a[]]"
	res, n, err := argFn.Parse(str)
	assert.NoError(t, err)
	vals := res.([]interface{})
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "xxx", vals[0].(string))
	a2vals := vals[1].([]interface{})
	assert.Equal(t, 0, len(a2vals))
	assert.Equal(t, len(str), n)

	str = "a[[]"
	_, _, err = argFn.Parse(str)
	assert.Equal(t, argparse.ErrFnUnclosedBracket, err)

	str = "a[]]"
	_, n, err = argFn.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, len(str)-1, n)
}
