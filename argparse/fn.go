package argparse

import (
	"fmt"
	"regexp"
)

var fnRe *regexp.Regexp

func init() {
	fnRe = regexp.MustCompile(`\A(\w+)(?:\[(.*)\])?`)
}

type ArgFn map[string]Fn

type Fn struct {
	Args Args
	Run  func([]interface{}) (interface{}, error)
}

func (a ArgFn) Parse(str string) (res interface{}, nparsed int, err error) {
	m := fnRe.FindStringSubmatch(str)
	if m == nil {
		err = ErrFnInvalidSyntax
		return
	}
	name, argsStr := m[1], m[2]
	fn, ok := a[name]
	if !ok {
		err = fmt.Errorf("no such function: %q", name)
		return
	}
	args, _, err := fn.Args.Parse(argsStr)
	if err != nil {
		return
	}
	nparsed = len(m[0])
	res, err = fn.Run(args.([]interface{}))
	return
}
