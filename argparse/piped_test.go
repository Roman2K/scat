package argparse_test

import (
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/Roman2K/scat/argparse"
)

func TestArgPiped(t *testing.T) {
	argVariadicTest{
		newArg: func(arg argparse.Parser) argparse.Parser {
			return argparse.ArgPiped{Arg: arg}
		},
		newStr: func(strs []string) string {
			return strings.Join(strs, string(argparse.Pipe))
		},
	}.run(t)
}

func TestArgPipedNested(t *testing.T) {
	arg := argparse.ArgPiped{
		Arg:  argparse.ArgStr,
		Nest: argparse.Brackets{'{', '}'},
	}
	str := "a | {b|c} | {d} | e | f{g|h}"
	res, n, err := arg.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, len(str), n)
	vals := res.([]interface{})
	assert.Equal(t, 5, len(vals))
	assert.Equal(t, "a", vals[0])
	assert.Equal(t, "{b|c}", vals[1])
	assert.Equal(t, "{d}", vals[2])
	assert.Equal(t, "e", vals[3])
	assert.Equal(t, "f{g|h}", vals[4])

	str = "a | {b|c"
	res, n, err = arg.Parse(str)
	assert.NoError(t, err)
	assert.Equal(t, len(str), n)
	vals = res.([]interface{})
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "a", vals[0])
	assert.Equal(t, "{b|c", vals[1])
}
