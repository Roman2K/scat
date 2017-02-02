package argparse_test

import (
	"errors"
	"testing"

	"gitlab.com/Roman2K/scat/argparse"
	assert "github.com/stretchr/testify/require"
)

func TestArgLambdaArgErr(t *testing.T) {
	arg := argparse.ArgLambda{
		Args: argparse.Args{argErr{errors.New("some err")}},
		Run: func([]interface{}) (interface{}, error) {
			return nil, nil
		},
	}
	_, _, err := arg.Parse("[a]")
	assert.Regexp(t, "some err", err)
}

func TestArgLambdaNested(t *testing.T) {
	arg1 := argparse.ArgLambda{
		Args: argparse.Args{argparse.ArgStr},
		Run: func(args []interface{}) (interface{}, error) {
			return args[0], nil
		},
	}
	arg2 := argparse.ArgLambda{
		Args: argparse.Args{argparse.ArgVariadic{arg1}},
		Run: func(args []interface{}) (interface{}, error) {
			return args[0], nil
		},
	}

	str := "[[a] [b]]"
	res, n, err := arg2.Parse(str)
	assert.NoError(t, err)
	vals := res.([]interface{})
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "a", vals[0].(string))
	assert.Equal(t, "b", vals[1].(string))
	assert.Equal(t, len(str), n)

	str = "[[]"
	_, _, err = arg2.Parse(str)
	assert.Equal(t, argparse.ErrUnclosedBracket, err)

	str = "[]]"
	res, n, err = arg2.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 0, len(vals))
	assert.Equal(t, 2, n)
}

func TestArgLambdaWrongArgsType(t *testing.T) {
	arg := argparse.ArgLambda{
		Args: argparse.ArgStr,
	}
	_, _, err := arg.Parse("[a]")
	expected := `Args parser.*is string.* not \[\]interface {}`
	assert.Regexp(t, expected, err.Error())
}
