package argparse_test

import (
	"strings"
	"testing"

	"gitlab.com/Roman2K/scat/argparse"
)

func TestArgPiped(t *testing.T) {
	argVariadicTest{
		newArg: func(arg argparse.Parser) argparse.Parser {
			return argparse.ArgPiped{arg}
		},
		newStr: func(strs []string) string {
			return strings.Join(strs, string(argparse.Pipe))
		},
	}.run(t)
}
