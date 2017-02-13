package argparse

type ArgVariadic struct {
	Arg Parser
}

func (ArgVariadic) Empty() (interface{}, error) {
	return []interface{}{}, nil
}

func (arg ArgVariadic) Parse(str string) (interface{}, int, error) {
	values := []interface{}{}
	pos, inLen := 0, len(str)
	for {
		pos += countLeftSpaces(str[pos:])
		if pos >= inLen {
			break
		}
		val, n, err := arg.Arg.Parse(str[pos:])
		pos += n
		if err != nil {
			err = ErrDetails{err, str, pos}
			return nil, pos, err
		}
		values = append(values, val)
	}
	return values, pos, nil
}
