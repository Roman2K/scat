package argparse

import "errors"

var (
	ErrTooManyArgs = errors.New("too many args")
	ErrTooFewArgs  = errors.New("too few args")
)

type Args []Parser

var _ Parser = Args{}

func (args Args) Parse(str string) (res interface{}, nparsed int, err error) {
	values := make([]interface{}, len(args))
	inLen := len(str)
	for i, arg := range args {
		nparsed += countLeftSpaces(str[nparsed:])
		if nparsed >= inLen {
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
		val, n, e := arg.Parse(str[nparsed:])
		nparsed += n
		if e != nil {
			err = errDetails(e, str, nparsed)
			return
		}
		values[i] = val
	}
	nparsed += countLeftSpaces(str[nparsed:])
	res = values
	if nparsed != inLen {
		err = ErrTooManyArgs
	}
	return
}
