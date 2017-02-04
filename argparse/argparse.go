package argparse

import (
	"errors"
	"fmt"
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

type ErrDetails struct {
	Err     error
	Str     string
	NParsed int
}

func (e ErrDetails) Error() string {
	msg := e.Err.Error()
	if strings.Contains(msg, "\n") {
		return msg
	}
	return fmt.Sprintf("%s\n  in \"%s\"\n%*s^", msg, e.Str, e.NParsed+6, " ")
}
