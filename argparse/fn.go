package argparse

import (
	"fmt"
	"regexp"
)

var fnRe *regexp.Regexp

func init() {
	var (
		open  = regexp.QuoteMeta(string(lambdaBrackets.Open))
		close = regexp.QuoteMeta(string(lambdaBrackets.Close))
	)
	fnRe = regexp.MustCompile(`\A(\w+)((?s)` + open + `.*` + close + `)?`)
}

type ArgFn map[string]Parser

func (a ArgFn) Parse(str string) (res interface{}, pos int, err error) {
	m := fnRe.FindStringSubmatch(str)
	if m == nil {
		err = ErrInvalidSyntax
		return
	}
	name, argsStr := m[1], m[2]
	parser, ok := a[name]
	if !ok {
		err = fmt.Errorf("no such function: %q", name)
		return
	}
	pos = len(name)
	posAdjust := 0
	if len(argsStr) == 0 {
		pos += countLeftSpaces(str[pos:])
		open, close := string(lambdaBrackets.Open), string(lambdaBrackets.Close)
		argsStr = open + str[pos:] + close
		posAdjust -= 2
	}
	var n int
	res, n, err = parser.Parse(argsStr)
	pos += n
	if err != nil {
		err = ErrDetails{err, str, pos}
		return
	}
	pos += posAdjust
	return
}
