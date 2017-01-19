package argparse_test

import (
	"errors"
	"strconv"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat/argparse"
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
	assert.Error(t, err)
	assert.IsType(t, &strconv.NumError{}, err)
}

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

	// arg err
	argFn = argparse.ArgFn{
		"a": argparse.Fn{
			Args: argparse.Args{argErr{someErr}},
			Run: func([]interface{}) (interface{}, error) {
				return nil, nil
			},
		},
	}
	_, _, err = argFn.Parse("a[b]")
	assert.Equal(t, someErr, err)
}

func TestArgVariadic(t *testing.T) {
	arg := argparse.ArgVariadic{argparse.ArgStr}

	str := ""
	res, n, err := arg.Parse(str)
	assert.NoError(t, err)
	vals := res.([]interface{})
	assert.Equal(t, 0, len(vals))
	assert.Equal(t, 0, n)

	str = "x"
	res, n, err = arg.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 1, len(vals))
	assert.Equal(t, 1, n)
	assert.Equal(t, "x", vals[0].(string))

	str = "x y"
	res, n, err = arg.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, 3, n)
	assert.Equal(t, "x", vals[0].(string))
	assert.Equal(t, "y", vals[1].(string))

	str = " x y "
	res, n, err = arg.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, 5, n)
	assert.Equal(t, "x", vals[0].(string))
	assert.Equal(t, "y", vals[1].(string))

	// arg err
	someErr := errors.New("some err")
	arg = argparse.ArgVariadic{argErr{someErr}}
	_, _, err = arg.Parse("x")
	assert.Equal(t, someErr, err)
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

type argErr struct {
	err error
}

func (a argErr) Parse(string) (interface{}, int, error) {
	return nil, 0, a.err
}

type argEmptyErr struct {
	err error
}

func (argEmptyErr) Parse(string) (interface{}, int, error) {
	return nil, 0, nil
}

func (a argEmptyErr) Empty() (interface{}, error) {
	return nil, a.err
}
