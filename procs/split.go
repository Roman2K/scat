package procs

import ss "secsplit"

func Split() Proc {
	return procFunc(split)
}

func split(c *ss.Chunk) Res {
	const num = 2
	offset := len(c.Data) / num
	boundaries := [num][2]int{{0, offset}, {offset, len(c.Data)}}
	chunks := make([]*ss.Chunk, len(boundaries))
	for i, bds := range boundaries {
		start, end := bds[0], bds[1]
		data := make([]byte, end-start)
		copy(data, c.Data[start:end])
		// TODO check overflow
		chunks[i] = &ss.Chunk{Num: c.Num*num + i, Data: data}
	}
	return Res{Chunks: chunks}
}
