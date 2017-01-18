package argparse

import "strconv"

var ArgInt = argInt{}

type argInt struct{}

func (argInt) Parse(str string) (interface{}, int, error) {
	i := spaceEndIndex(str)
	n, err := strconv.Atoi(str[:i])
	return n, i, err
}
