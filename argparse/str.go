package argparse

var ArgStr = argStr{}

type argStr struct{}

func (argStr) Parse(str string) (interface{}, int, error) {
	i := spaceEndIndex(str)
	return str[:i], i, nil
}
