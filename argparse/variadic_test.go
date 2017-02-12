package argparse_test

import (
	"errors"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"
	ap "gitlab.com/Roman2K/scat/argparse"
)

func TestArgVariadic(t *testing.T) {
	argVariadicTest{
		newArg: func(arg ap.Parser) ap.Parser {
			return ap.ArgVariadic{arg}
		},
		newStr: func(strs []string) string {
			return strings.Join(strs, " ")
		},
	}.run(t)
}

type argVariadicTest struct {
	newArg func(ap.Parser) ap.Parser
	newStr func([]string) string
}

func (test argVariadicTest) run(t *testing.T) {
	arg := test.newArg(ap.ArgStr)

	str := test.newStr([]string{""})
	res, n, err := arg.Parse(str)
	assert.NoError(t, err)
	vals := res.([]interface{})
	assert.Equal(t, 0, len(vals))
	assert.Equal(t, len(str), n)

	str = test.newStr([]string{"x"})
	res, n, err = arg.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 1, len(vals))
	assert.Equal(t, len(str), n)
	assert.Equal(t, "x", vals[0].(string))

	str = test.newStr([]string{"x", "y"})
	res, n, err = arg.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, len(str), n)
	assert.Equal(t, "x", vals[0].(string))
	assert.Equal(t, "y", vals[1].(string))

	str = test.newStr([]string{" x ", " y "})
	res, n, err = arg.Parse(str)
	assert.NoError(t, err)
	vals = res.([]interface{})
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, len(str), n)
	assert.Equal(t, "x", vals[0].(string))
	assert.Equal(t, "y", vals[1].(string))

	// arg err
	someErr := errors.New("some err")
	arg = ap.ArgVariadic{argErr{someErr}}
	_, _, err = arg.Parse("x")
	assert.Equal(t, someErr, err.(ap.ErrDetails).Err)

	// err nparsed
	someErr = errors.New("some err")
	strs := []string{"xxx", "yyy"}
	str = test.newStr(strs)
	arg = test.newArg(&argFailAfter{
		arg:     ap.ArgStr,
		after:   1,
		nparsed: 2,
		err:     someErr,
	})
	_, _, err = arg.Parse(str)
	errDet, ok := err.(ap.ErrDetails)
	assert.True(t, ok)
	assert.Equal(t, someErr, errDet.Err)
	assert.Equal(t, len(str)-len(strs[1])+2, errDet.NParsed)
}

type argFailAfter struct {
	arg     ap.Parser
	after   uint
	nparsed int
	err     error
	cur     uint
}

func (arg *argFailAfter) Parse(str string) (val interface{}, n int, err error) {
	arg.cur++
	val, n, err = arg.arg.Parse(str)
	if arg.cur > arg.after {
		n = arg.nparsed
		err = arg.err
	}
	return
}
