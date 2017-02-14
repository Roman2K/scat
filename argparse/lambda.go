package argparse

import (
	"errors"
	"fmt"
	"strings"
)

var (
	lambdaBrackets = Brackets{'(', ')'}
)

var (
	ErrUnclosedBracket = errors.New("unclosed bracket")
)

type ArgLambda struct {
	Args     Parser
	Run      RunFn
	Brackets Brackets
}

type RunFn func([]interface{}) (interface{}, error)

func (a ArgLambda) Parse(str string) (res interface{}, pos int, err error) {
	brackets := lambdaBrackets
	if a.Brackets != noBrackets {
		brackets = a.Brackets
	}
	if len(str) < 2 || []rune(str)[0] != brackets.Open {
		err = ErrInvalidSyntax
		return
	}
	str = str[1:]
	nest := 1
	i := strings.IndexFunc(str, func(r rune) bool {
		switch r {
		case brackets.Open:
			nest++
		case brackets.Close:
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
	pos = i + 2
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
