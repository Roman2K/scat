package argparse

import "strings"

type ArgPiped struct {
	Arg  Parser
	Nest Brackets
}

type Brackets struct {
	Open, Close rune
}

var noBrackets = Brackets{}

const Pipe = '|'

func (arg ArgPiped) Parse(str string) (interface{}, int, error) {
	values := make([]interface{}, 0, strings.Count(str, string(Pipe))+1)
	pos, inLen := 0, len(str)
	for {
		pos += countLeftSpaces(str[pos:])
		if pos >= inLen {
			break
		}
		argStr := str[pos:]
		match := func(r rune) bool {
			return r == Pipe
		}
		if arg.Nest != noBrackets {
			nest := 0
			match = func(r rune) bool {
				switch r {
				case arg.Nest.Open:
					nest++
				case arg.Nest.Close:
					nest--
				case Pipe:
					if nest == 0 {
						return true
					}
				}
				return false
			}
		}
		posAdjust := 0
		if i := strings.IndexFunc(argStr, match); i != -1 {
			argStr = argStr[:i]
			posAdjust += 1
		}
		val, n, err := arg.Arg.Parse(argStr)
		if err != nil {
			err = ErrDetails{err, str, pos + n}
			return nil, pos, err
		}
		pos += len(argStr) + posAdjust
		values = append(values, val)
	}
	return values, pos, nil
}
