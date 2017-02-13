package argparse

import "strings"

type ArgPair struct {
	Left  Parser
	Right Parser
	Run   func(_, _ interface{}) (interface{}, error)
}

const pairSep = '='

func (arg ArgPair) Parse(str string) (res interface{}, nparsed int, err error) {
	skipSep := 0
	if _, ok := arg.Left.(ArgPair); ok {
		skipSep++
	}
	i := strings.IndexFunc(str, func(r rune) bool {
		if r != pairSep {
			return false
		}
		if skipSep <= 0 {
			return true
		}
		skipSep--
		return false
	})
	if i == -1 {
		err = ErrInvalidSyntax
		return
	}
	left, n, err := arg.Left.Parse(str[:i])
	if err != nil {
		err = ErrDetails{err, str, nparsed + n}
		return
	}
	nparsed += i + 1
	right, n, err := arg.Right.Parse(str[nparsed:])
	nparsed += n
	if err != nil {
		err = ErrDetails{err, str, nparsed}
		return
	}
	res, err = arg.Run(left, right)
	return
}
