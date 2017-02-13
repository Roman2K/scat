package argparse_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"
	"gitlab.com/Roman2K/scat/argparse"
)

func TestArgPair(t *testing.T) {
	type pair struct {
		a, b interface{}
	}

	arg1 := argparse.ArgPair{
		Left:  argparse.ArgStr,
		Right: argparse.ArgStr,
		Run: func(a, b interface{}) (interface{}, error) {
			return pair{a, b}, nil
		},
	}
	arg2 := argparse.ArgPair{
		Left:  arg1,
		Right: argparse.ArgStr,
		Run: func(a, b interface{}) (interface{}, error) {
			return pair{a, b}, nil
		},
	}
	arg3 := argparse.ArgPair{
		Left:  argparse.ArgStr,
		Right: arg1,
		Run: func(a, b interface{}) (interface{}, error) {
			return pair{a, b}, nil
		},
	}

	str := "a=b x "
	res, n, err := arg1.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	val := res.(pair)
	assert.Equal(t, "a", val.a.(string))
	assert.Equal(t, "b", val.b.(string))

	str = "a= "
	res, n, err = arg1.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	val = res.(pair)
	assert.Equal(t, "a", val.a.(string))
	assert.Equal(t, "", val.b.(string))

	str = "a x"
	_, n, err = arg1.Parse(str)
	assert.Equal(t, argparse.ErrInvalidSyntax, err)
	assert.Equal(t, 0, n)

	str = "a=b=c x "
	res, n, err = arg2.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	val = res.(pair)
	assert.Equal(t, "a", val.a.(pair).a.(string))
	assert.Equal(t, "b", val.a.(pair).b.(string))
	assert.Equal(t, "c", val.b.(string))

	str = "a=b=c x "
	res, n, err = arg3.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	val = res.(pair)
	assert.Equal(t, "a", val.a.(string))
	assert.Equal(t, "b", val.b.(pair).a.(string))
	assert.Equal(t, "c", val.b.(pair).b.(string))
}
