package argparse

import (
	"fmt"
	"regexp"
)

var fnRe *regexp.Regexp

func init() {
	var (
		open  = regexp.QuoteMeta(string(lambdaOpen))
		close = regexp.QuoteMeta(string(lambdaClose))
	)
	fnRe = regexp.MustCompile(`\A(\w+)(` + open + `.*` + close + `)?`)
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
	nparsedAdjust := 0
	if len(argsStr) == 0 {
		nparsed += countLeftSpaces(str[nparsed:])
		argsStr = string(lambdaOpen) + str[nparsed:] + string(lambdaClose)
		nparsedAdjust -= 2
	}
	var n int
	res, n, err = parser.Parse(argsStr)
	nparsed += n
	if err != nil {
		err = ErrDetails{err, str, nparsed}
		return
	}
	nparsed += nparsedAdjust
	return
}
