package argparse

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrFnInvalidSyntax   = errors.New("invalid syntax for function arg")
	ErrFnUnclosedBracket = errors.New("unclosed bracket")
	ErrFnUnopenedBracket = errors.New("unopened bracket")
)

var fnRe *regexp.Regexp

func init() {
	fnRe = regexp.MustCompile(`\A(\w+)(\[.*\])?`)
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
	nparsed = len(name)
	nest := 0
	for i, r := range argsStr {
		if r == '[' {
			nest++
		} else if r == ']' {
			nest--
			if nest == 0 {
				argsStr = argsStr[1:i]
				nparsed += len(argsStr) + 2
				break
			}
		}
	}
	if nest > 0 {
		err = ErrFnUnclosedBracket
		return
	}
	args, _, err := fn.Args.Parse(argsStr)
	if err != nil {
		return
	}
	res, err = fn.Run(args.([]interface{}))
	return
}
