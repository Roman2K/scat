package argparse

import (
	"fmt"
	"regexp"
)

var fnRe *regexp.Regexp

func init() {
	fnRe = regexp.MustCompile(`\A(\w+)(\[.*\])?`)
}

type ArgFn map[string]Parser

func (a ArgFn) Parse(str string) (res interface{}, nparsed int, err error) {
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
	nparsed = len(name)
	if argsStr == "" {
		argsStr = "[]"
		nparsed -= len(argsStr)
	}
	var n int
	res, n, err = parser.Parse(argsStr)
	nparsed += n
	return
}
