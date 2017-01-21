package argparse

import (
	"errors"
	"strings"
)

var (
	ErrUnclosedBracket = errors.New("unclosed bracket")
)

var (
	lambdaOpen  = '['
	lambdaClose = ']'
)

type ArgLambda struct {
	Args Args
	Run  func([]interface{}) (interface{}, error)
}

func (a ArgLambda) Parse(str string) (res interface{}, nparsed int, err error) {
	if len(str) < 2 || []rune(str)[0] != lambdaOpen {
		err = ErrInvalidSyntax
		return
	}
	str = str[1:]
	nest := 1
	i := strings.IndexFunc(str, func(r rune) bool {
		switch r {
		case lambdaOpen:
			nest++
		case lambdaClose:
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
	args, _, err := a.Args.Parse(str)
	if err != nil {
		return
	}
	res, err = a.Run(args.([]interface{}))
	nparsed = i + 2
	return
}
