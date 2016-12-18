package procs

import ss "secsplit"

var Size Proc

func init() {
	Size = inplaceProcFunc(setSize)
}

func setSize(c *ss.Chunk) error {
	c.Size = len(c.Data)
	return nil
}
