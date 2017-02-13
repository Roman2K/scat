package argparse

import (
	"strconv"

	humanize "github.com/dustin/go-humanize"
)

var ArgBytes = argBytes{}

type argBytes struct{}

func (argBytes) Parse(str string) (interface{}, int, error) {
	i := spaceEndIndex(str)
	n, err := humanize.ParseBytes(str[:i])
	if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrSyntax {
		err = ErrInvalidSyntax
	}
	return n, i, err
}
