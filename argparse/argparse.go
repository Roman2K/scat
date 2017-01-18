package argparse

import (
	"errors"
	"strings"
	"unicode"
)

var (
	ErrFnInvalidSyntax = errors.New("invalid syntax for function arg")
	ErrTooManyArgs     = errors.New("too many args")
	ErrTooFewArgs      = errors.New("too few args")
)

type Parser interface {
	Parse(string) (interface{}, int, error)
}

func countLeftSpaces(str string) int {
	trimmed := strings.TrimLeftFunc(str, unicode.IsSpace)
	return len(str) - len(trimmed)
}

func spaceEndIndex(str string) int {
	if i := strings.IndexFunc(str, unicode.IsSpace); i != -1 {
		return i
	}
	return len(str)
}
