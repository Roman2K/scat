package argparse

import "strings"

type ArgPair struct {
	Left  Parser
	Right Parser
	Run   func(_, _ interface{}) (interface{}, error)
}

const pairSep = '='

func (arg ArgPair) Parse(str string) (res interface{}, pos int, err error) {
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
		err = ErrDetails{err, str, pos + n}
		return
	}
	pos += i + 1
	right, n, err := arg.Right.Parse(str[pos:])
	pos += n
	if err != nil {
		err = ErrDetails{err, str, pos}
		return
	}
	res, err = arg.Run(left, right)
	return
}
