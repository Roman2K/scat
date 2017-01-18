package argparse

import humanize "github.com/dustin/go-humanize"

var ArgBytes = argBytes{}

type argBytes struct{}

func (argBytes) Parse(str string) (interface{}, int, error) {
	i := spaceEndIndex(str)
	n, err := humanize.ParseBytes(str[:i])
	return n, i, err
}
