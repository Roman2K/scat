package stores

type StrPart []int

func (nest StrPart) Split(str string) (parts []string) {
	parts = make([]string, len(nest))
	offset := 0
	nchars := len(str)
	for i, n := range nest {
		if left := nchars - offset; n > left {
			n = left
		}
		parts[i] = str[offset : offset+n]
		offset += n
	}
	return
}
