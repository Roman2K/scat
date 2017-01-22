package argparse

type ArgFilter struct {
	Parser Parser
	Filter func(interface{}) (interface{}, error)
}

func (a ArgFilter) Parse(str string) (res interface{}, n int, err error) {
	res, n, err = a.Parser.Parse(str)
	if err != nil {
		return
	}
	res, err = a.Filter(res)
	return
}
