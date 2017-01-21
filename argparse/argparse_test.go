package argparse_test

type argErr struct {
	err error
}

func (a argErr) Parse(string) (interface{}, int, error) {
	return nil, 0, a.err
}

type argEmptyErr struct {
	err error
}

func (argEmptyErr) Parse(string) (interface{}, int, error) {
	return nil, 0, nil
}

func (a argEmptyErr) Empty() (interface{}, error) {
	return nil, a.err
}
