package argparse

type ArgOr []Parser

func (args ArgOr) Parse(str string) (res interface{}, n int, err error) {
	for _, arg := range args {
		res, n, err = arg.Parse(str)
		if OriginalErr(err) == ErrInvalidSyntax {
			continue
		}
		return
	}
	return
}
