package argparse

import "errors"

var (
	ErrTooManyArgs = errors.New("too many args")
	ErrTooFewArgs  = errors.New("too few args")
)

type Args []Parser

var _ Parser = Args{}

func (args Args) Parse(str string) (res interface{}, pos int, err error) {
	values := make([]interface{}, len(args))
	inLen := len(str)
	for i, arg := range args {
		pos += countLeftSpaces(str[pos:])
		if pos >= inLen {
			if ep, ok := arg.(EmptyParser); ok {
				val, e := ep.Empty()
				if e != nil {
					err = e
					return
				}
				values[i] = val
				continue
			}
			err = ErrTooFewArgs
			return
		}
		val, n, e := arg.Parse(str[pos:])
		pos += n
		if e != nil {
			err = ErrDetails{e, str, pos}
			return
		}
		values[i] = val
	}
	pos += countLeftSpaces(str[pos:])
	res = values
	if pos != inLen {
		err = ErrTooManyArgs
	}
	return
}
