package argparse

import "fmt"

type ArgOr []Parser

func (args ArgOr) Parse(str string) (res interface{}, n int, err error) {
	for _, arg := range args {
		res, n, err = arg.Parse(str)
		if err == nil {
			return
		}
	}
	err = fmt.Errorf("no parser matched, last err: %v", err)
	return
}
