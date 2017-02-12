package argparse

import (
	"errors"
	"fmt"
	"strings"
)

const (
	lambdaOpen  = '('
	lambdaClose = ')'
)

var (
	ErrUnclosedBracket = errors.New("unclosed bracket")
)

type ArgLambda struct {
	Args  Parser
	Run   RunFn
	Open  rune
	Close rune
}

type RunFn func([]interface{}) (interface{}, error)

func (a ArgLambda) Parse(str string) (res interface{}, nparsed int, err error) {
	open, close := lambdaOpen, lambdaClose
	if a.Open != 0 && a.Close != 0 {
		open, close = a.Open, a.Close
	}
	if len(str) < 2 || []rune(str)[0] != open {
		err = ErrInvalidSyntax
		return
	}
	str = str[1:]
	nest := 1
	i := strings.IndexFunc(str, func(r rune) bool {
		switch r {
		case open:
			nest++
		case close:
			nest--
			if nest == 0 {
				return true
			}
		}
		return false
	})
	if nest > 0 {
		err = ErrUnclosedBracket
		return
	}
	str = str[:i]
	iargs, _, err := a.args().Parse(str)
	if err != nil {
		return
	}
	args, argsErr := func() (res []interface{}, err interface{}) {
		defer func() {
			err = recover()
		}()
		res = iargs.([]interface{})
		return
	}()
	if argsErr != nil {
		err = fmt.Errorf("wrong return type of Args parser: %v", argsErr)
		return
	}
	res, err = a.run()(args)
	nparsed = i + 2
	return
}

func (a ArgLambda) args() Parser {
	if a.Args == nil {
		return Args{}
	}
	return a.Args
}

func (a ArgLambda) run() RunFn {
	if a.Run == nil {
		return nopRun
	}
	return a.Run
}

var nopRun RunFn = func([]interface{}) (interface{}, error) {
	return nil, nil
}
