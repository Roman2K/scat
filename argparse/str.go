package argparse

import (
	"strings"
	"unicode"
)

var ArgStr = argStr{}

type argStr struct{}

func (argStr) Parse(str string) (interface{}, int, error) {
	r := strings.NewReader(str)
	first, bytesUsed, err := r.ReadRune()
	if err != nil {
		return "", 0, nil
	}
	if first != '"' {
		// Not a quoted string; seek the next space.
		return spaceOnlyParse(str)
	}

	// Search for a terminating quote.
	collectedRunes := []rune{}
	totalBytesConsumed := bytesUsed
	for {
		current, bytesUsed, err := r.ReadRune()
		if err != nil {
			// Malformed or not a quoted string; seek the next space.
			return spaceOnlyParse(str)
		}
		totalBytesConsumed += bytesUsed
		if current == '"' {
			// Before accepting, confirm that the following
			// character is whitespace or EOF, to reduce the
			// risk of changing behavior of an old string.
			next, _, err := r.ReadRune()
			if err == nil && !unicode.IsSpace(next) {
				return spaceOnlyParse(str)
			}
			return string(collectedRunes), totalBytesConsumed, nil
		}
		collectedRunes = append(collectedRunes, current)
	}
}

func spaceOnlyParse(str string) (interface{}, int, error) {
	i := spaceEndIndex(str)
	return str[:i], i, nil
}
