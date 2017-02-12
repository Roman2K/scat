package argparse

import "strings"

type ArgPiped struct {
	Arg Parser
}

const Pipe = '|'

func (arg ArgPiped) Parse(str string) (interface{}, int, error) {
	values := make([]interface{}, 0, strings.Count(str, string(Pipe))+1)
	nparsed, inLen := 0, len(str)
	for {
		nparsed += countLeftSpaces(str[nparsed:])
		if nparsed >= inLen {
			break
		}
		argStr := str[nparsed:]
		nparsedAdjust := 0
		if i := strings.IndexRune(str[nparsed:], Pipe); i != -1 {
			argStr = argStr[:i]
			nparsedAdjust += len(string(Pipe))
		}
		val, n, err := arg.Arg.Parse(argStr)
		if err != nil {
			err = ErrDetails{err, str, nparsed + n}
			return nil, nparsed, err
		}
		nparsed += len(argStr) + nparsedAdjust
		values = append(values, val)
	}
	return values, nparsed, nil
}
