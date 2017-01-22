package argparse

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

var (
	ErrInvalidSyntax = errors.New("invalid syntax")
)

type Parser interface {
	Parse(string) (interface{}, int, error)
}

type EmptyParser interface {
	Empty() (interface{}, error)
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

func errDetails(err error, parser interface{}, str string, n int) error {
	pt := strings.TrimPrefix(reflect.TypeOf(parser).Name(), "Arg")
	if strings.Count(err.Error(), "\n") > 0 {
		return fmt.Errorf("%s: %v", pt, err)
	}
	return fmt.Errorf("%s: %v\n  in \"%s\"\n%*s^", pt, err, str, n+6, " ")
}
